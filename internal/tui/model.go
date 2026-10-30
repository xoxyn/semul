package tui

import (
	"context"
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/xoxyn/semul/internal/domain"
	"github.com/xoxyn/semul/internal/reviewer"
)

type mode int

const (
	modeList mode = iota
	modeEdit
	modeConfirm
	modeHelp
)

type sentMsg struct {
	idx int
	err error
}

// SendFunc delivers one approved email. It runs in its own goroutine (as a
// tea.Cmd), so it is fine for it to block.
type SendFunc func(ctx context.Context, job domain.EmailJob) error

type Model struct {
	jobs []domain.EmailJob
	send SendFunc
	// queue holds indices of jobs waiting to be sent; they go out one at a time.
	queue []int

	cursor        int
	scroll        int
	previewScroll int

	width, height int
	mode          mode

	ed      editor
	editIdx int

	log []string
}

// New builds the review screen for the given drafts. send is called once per
// approved email when the user confirms sending.
func New(jobs []domain.EmailJob, send SendFunc) Model {
	if send == nil {
		send = func(context.Context, domain.EmailJob) error {
			return errors.New("no sender configured")
		}
	}
	m := Model{jobs: jobs, send: send}
	m.logf("loaded %d drafts", len(jobs))
	return m
}

// Run starts the TUI and blocks until the user quits.
func Run(jobs []domain.EmailJob, send SendFunc) error {
	_, err := tea.NewProgram(New(jobs, send)).Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) logf(format string, a ...any) {
	line := c("90", time.Now().Format("15:04:05")) + " " + fmt.Sprintf(format, a...)
	m.log = append(m.log, line)
	if len(m.log) > 100 {
		m.log = m.log[len(m.log)-100:]
	}
}

// layout returns widths/heights: left width, right width, drafts height, log height, content height.
func (m Model) layout() (lw, rw, dh, lh, ch int) {
	ch = m.height - 1
	lw = max(32, m.width*35/100)
	rw = m.width - lw
	lh = max(5, ch/3)
	dh = ch - lh
	return
}

func (m *Model) clamp() {
	_, _, dh, _, _ := m.layout()
	vis := dh - 2
	if vis < 1 {
		return
	}
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+vis {
		m.scroll = m.cursor - vis + 1
	}
}

func (m *Model) scrollPreview(delta int) {
	_, _, _, _, ch := m.layout()
	maxScroll := max(0, len(m.previewLines())-(ch-2))
	m.previewScroll = max(0, min(m.previewScroll+delta, maxScroll))
}

func (m *Model) syncEditor() {
	if len(m.ed.body) == 0 {
		return // no editor open yet (e.g. first WindowSizeMsg)
	}
	m.ed.w = max(1, m.width-4)
	m.ed.h = max(1, m.height-1-3-2)
	m.ed.fixScroll()
}

func (m *Model) move(to int) {
	if len(m.jobs) == 0 {
		return
	}
	m.cursor = max(0, min(to, len(m.jobs)-1))
	m.previewScroll = 0
	m.clamp()
}

func (m Model) pendingSend() []int {
	var idx []int
	for i, j := range m.jobs {
		if j.Approved && j.Status != domain.Sent && j.Status != domain.Sending {
			idx = append(idx, i)
		}
	}
	return idx
}

// sendNext pops the next queued job and returns a Cmd that sends it. The
// result comes back to Update as a sentMsg, which then starts the next one.
func (m *Model) sendNext() tea.Cmd {
	if len(m.queue) == 0 {
		return nil
	}
	i := m.queue[0]
	m.queue = m.queue[1:]

	job, send := m.jobs[i], m.send // copy: the goroutine must not touch m
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return sentMsg{idx: i, err: send(ctx, job)}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clamp()
		m.syncEditor()
		return m, nil

	case sentMsg:
		j := &m.jobs[msg.idx]
		if msg.err != nil {
			j.Status = domain.Failed
			j.Approved = false
			m.logf("%s %s: %v", c("31", "failed"), j.Company, msg.err)
		} else {
			j.Status = domain.Sent
			m.logf("%s %s", c("36", "sent"), j.Company)
		}
		cmd := m.sendNext() // start the next queued email, if any
		return m, cmd

	case tea.PasteMsg:
		if m.mode == modeEdit {
			m.ed.insert(msg.Content)
			m.syncEditor()
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.mode {
		case modeList:
			return m.updateList(msg)
		case modeEdit:
			return m.updateEdit(msg)
		case modeConfirm:
			return m.updateConfirm(msg)
		case modeHelp:
			m.mode = modeList
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	_, _, _, _, ch := m.layout()
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "j", "down":
		m.move(m.cursor + 1)
	case "k", "up":
		m.move(m.cursor - 1)
	case "g", "home":
		m.move(0)
	case "G", "end":
		m.move(len(m.jobs) - 1)

	case "ctrl+d", "pgdown":
		m.scrollPreview((ch - 2) / 2)
	case "ctrl+u", "pgup":
		m.scrollPreview(-(ch - 2) / 2)

	case "space", " ", "a":
		if len(m.jobs) == 0 {
			break
		}
		j := &m.jobs[m.cursor]
		if j.Status == domain.Sent || j.Status == domain.Sending {
			m.logf("%s already sent", j.Company)
			break
		}
		reviewer.Toggle(j)
		if j.Approved {
			m.logf("approved %s", j.Company)
		} else {
			m.logf("unapproved %s", j.Company)
		}

	case "A":
		n := 0
		for i := range m.jobs {
			j := &m.jobs[i]
			if !j.Approved && j.Status != domain.Sent && j.Status != domain.Sending {
				reviewer.Approve(j)
				n++
			}
		}
		m.logf("approved %d drafts", n)

	case "e", "enter":
		if len(m.jobs) == 0 {
			break
		}
		j := m.jobs[m.cursor]
		if j.Status == domain.Sent || j.Status == domain.Sending {
			m.logf("%s already sent, can't edit", j.Company)
			break
		}
		m.editIdx = m.cursor
		m.ed = newEditor(j.Draft.Subject, j.Draft.Body)
		m.syncEditor()
		m.mode = modeEdit

	case "s":
		if len(m.pendingSend()) == 0 {
			m.logf("nothing approved to send")
			break
		}
		m.mode = modeConfirm

	case "?":
		m.mode = modeHelp
	}
	return m, nil
}

func (m Model) updateEdit(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
		m.logf("edit cancelled")
		return m, nil

	case "ctrl+s":
		subject, body := m.ed.text()
		j := &m.jobs[m.editIdx]
		if subject != j.Draft.Subject || body != j.Draft.Body {
			reviewer.Edit(j, subject, body) // also resets approval
			m.logf("edited %s (approval reset)", j.Company)
		} else {
			m.logf("no changes to %s", j.Company)
		}
		m.mode = modeList
		return m, nil

	case "tab", "shift+tab":
		m.ed.field = 1 - m.ed.field
		m.ed.want = -1
		if m.ed.field == 1 {
			m.ed.col = min(m.ed.col, len(m.ed.body[m.ed.row]))
		}

	default:
		m.ed.key(msg)
	}
	m.syncEditor()
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.queue = m.pendingSend()
		for _, i := range m.queue {
			m.jobs[i].Status = domain.Sending // queued
		}
		m.logf("sending %d emails...", len(m.queue))
		m.mode = modeList
		cmd := m.sendNext()
		return m, cmd
	case "n", "esc", "q":
		m.mode = modeList
	}
	return m, nil
}

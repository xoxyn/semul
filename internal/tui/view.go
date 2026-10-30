package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/xoxyn/semul/internal/domain"
)

func statusIcon(j domain.EmailJob) (string, string) {
	switch {
	case j.Status == domain.Sent:
		return "➜", "36"
	case j.Status == domain.Sending:
		return "…", "33"
	case j.Approved:
		return "✓", "32"
	case j.Status == domain.Failed:
		return "✗", "31"
	default:
		return "○", "90"
	}
}

func statusLabel(j domain.EmailJob) string {
	switch {
	case j.Status == domain.Sent:
		return c("36", "sent")
	case j.Status == domain.Sending:
		return c("33", "sending…")
	case j.Approved:
		return c("32", "approved")
	case j.Status == domain.Failed:
		return c("31", "failed (approve to retry)")
	default:
		return c("90", "pending review")
	}
}

func (m Model) row(i, w int) string {
	j := m.jobs[i]
	ic, col := statusIcon(j)
	if i == m.cursor {
		return c("48;5;24;1;97", plainPad(" "+ic+" "+j.Company, w))
	}
	return " " + c(col, ic) + " " + j.Company
}

func (m Model) previewLines() []string {
	_, rw, _, _, _ := m.layout()
	w := rw - 4
	if len(m.jobs) == 0 {
		return []string{" No drafts to review."}
	}
	j := m.jobs[m.cursor]
	lines := []string{
		" " + c("90", "Company ") + c("1", j.Company),
		" " + c("90", "To      ") + j.Email,
		" " + c("90", "Problem ") + j.IdentifiedProblem,
		" " + c("90", "Status  ") + statusLabel(j),
		" " + c("90", strings.Repeat("─", max(0, w))),
		" " + c("1;36", "Subject"),
	}
	for _, l := range wrap(j.Draft.Subject, w) {
		lines = append(lines, " "+c("1", l))
	}
	lines = append(lines, "")
	for _, l := range wrap(j.Draft.Body, w) {
		lines = append(lines, " "+l)
	}
	return lines
}

func hints(pairs ...[2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = c("36", p[0]) + " " + p[1]
	}
	return " " + strings.Join(parts, "  ")
}

func (m Model) statusBar() string {
	switch m.mode {
	case modeConfirm:
		return c("1;33", fmt.Sprintf(" Send %d approved email(s)? (y/n)", len(m.pendingSend())))
	case modeEdit:
		return hints([2]string{"ctrl+s", "save"}, [2]string{"esc", "cancel"}, [2]string{"tab", "switch field"})
	}
	left := hints(
		[2]string{"space", "approve"},
		[2]string{"e", "edit"},
		[2]string{"A", "approve all"},
		[2]string{"s", "send"},
		[2]string{"?", "help"},
		[2]string{"q", "quit"},
	)
	approved, sent := 0, 0
	for _, j := range m.jobs {
		switch {
		case j.Status == domain.Sent:
			sent++
		case j.Approved:
			approved++
		}
	}
	right := c("32", fmt.Sprintf("%d approved", approved)) + c("90", " · ") +
		c("36", fmt.Sprintf("%d sent", sent)) + c("90", " · ") +
		c("90", fmt.Sprintf("%d total ", len(m.jobs)))
	gap := m.width - visWidth(left) - visWidth(right)
	if gap < 1 {
		return left
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) viewMain() string {
	lw, rw, dh, lh, ch := m.layout()

	// Drafts
	var rows []string
	for i := m.scroll; i < len(m.jobs) && len(rows) < dh-2; i++ {
		rows = append(rows, m.row(i, lw-2))
	}
	footer := ""
	if len(m.jobs) > 0 {
		footer = fmt.Sprintf("%d of %d", m.cursor+1, len(m.jobs))
	}
	drafts := box("Drafts", footer, rows, lw, dh, true)

	// Log (newest at the bottom)
	logVis := lh - 2
	logLines := m.log
	if len(logLines) > logVis {
		logLines = logLines[len(logLines)-logVis:]
	}
	logBox := box("Log", "", logLines, lw, lh, false)

	// Preview
	all := m.previewLines()
	ph := ch - 2
	sc := max(0, min(m.previewScroll, max(0, len(all)-ph)))
	end := min(sc+ph, len(all))
	pFooter := ""
	if len(all) > ph {
		pFooter = fmt.Sprintf("%d of %d", sc+1, len(all)-ph+1)
	}
	preview := box("Preview", pFooter, all[sc:end], rw, ch, false)

	return joinH(drafts+"\n"+logBox, preview) + "\n" + m.statusBar()
}

func (m Model) viewEdit() string {
	ch := m.height - 1
	j := m.jobs[m.editIdx]

	subj := " " + m.ed.renderSubject(m.ed.w, m.ed.field == 0)
	subjBox := box("Subject", "", []string{subj}, m.width, 3, m.ed.field == 0)

	body := m.ed.renderBody(m.ed.w, m.ed.h, m.ed.field == 1)
	for i := range body {
		body[i] = " " + body[i]
	}
	bodyBox := box("Body · "+j.Company+" <"+j.Email+">", "", body, m.width, ch-3, m.ed.field == 1)

	return subjBox + "\n" + bodyBox + "\n" + m.statusBar()
}

func (m Model) viewHelp() string {
	k := func(key, desc string) string { return "  " + c("36", plainPad(key, 12)) + desc }
	h := func(s string) string { return c("1;32", s) }
	content := strings.Join([]string{
		h("Navigation"),
		k("j / k, ↑ / ↓", "move between drafts"),
		k("g / G", "first / last draft"),
		k("ctrl+d / ctrl+u", "scroll preview"),
		"",
		h("Actions"),
		k("space, a", "approve / unapprove"),
		k("A", "approve all"),
		k("e, enter", "edit draft (resets approval)"),
		k("s", "send approved emails"),
		"",
		h("Editor"),
		k("tab", "switch subject / body"),
		k("ctrl+s", "save"),
		k("esc", "cancel"),
		"",
		c("90", "press any key to close"),
	}, "\n")
	return center(box("Help", "", strings.Split(content, "\n"), 52, 21, true), m.width, m.height)
}

func (m Model) View() tea.View {
	var s string
	switch {
	case m.width == 0:
		s = ""
	case m.width < 60 || m.height < 14:
		s = "Terminal too small (need at least 60x14)."
	case m.mode == modeEdit:
		s = m.viewEdit()
	case m.mode == modeHelp:
		s = m.viewHelp()
	default:
		s = m.viewMain()
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

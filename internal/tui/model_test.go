package tui

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/xoxyn/semul/internal/domain"
)

func testJobs() []domain.EmailJob {
	return []domain.EmailJob{
		{
			Prospect: domain.Prospect{Company: "Acme", Email: "a@acme.test"},
			Draft:    domain.DraftEmail{Subject: "Hi Acme", Body: "Body A"},
			Status:   domain.NotSent,
		},
		{
			Prospect: domain.Prospect{Company: "Globex", Email: "g@globex.test"},
			Draft:    domain.DraftEmail{Subject: "Hi Globex", Body: "Body B"},
			Status:   domain.NotSent,
		},
	}
}

// press sends one key press to the model and returns the new model and cmd.
func press(m Model, key string) (Model, tea.Cmd) {
	next, cmd := m.Update(tea.KeyPressMsg{Code: rune(key[0]), Text: key})
	return next.(Model), cmd
}

func newTestModel(send SendFunc) Model {
	m := New(testJobs(), send)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return next.(Model)
}

func TestApproveToggle(t *testing.T) {
	m := newTestModel(nil)

	m, _ = press(m, "a")
	if !m.jobs[0].Approved {
		t.Fatal("expected first job approved")
	}

	m, _ = press(m, "a")
	if m.jobs[0].Approved {
		t.Fatal("expected first job unapproved")
	}
}

func TestMoveCursor(t *testing.T) {
	m := newTestModel(nil)

	m, _ = press(m, "j")
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}

	m, _ = press(m, "j") // already at the end
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 (clamped)", m.cursor)
	}

	m, _ = press(m, "k")
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
}

func TestApproveAll(t *testing.T) {
	m := newTestModel(nil)

	m, _ = press(m, "A")

	for i, j := range m.jobs {
		if !j.Approved {
			t.Errorf("job %d not approved", i)
		}
	}
}

func TestSendWithNothingApprovedDoesNotAskForConfirmation(t *testing.T) {
	m := newTestModel(nil)

	m, _ = press(m, "s")

	if m.mode != modeList {
		t.Errorf("mode = %v, want modeList", m.mode)
	}
}

func TestEditSaveResetsApproval(t *testing.T) {
	m := newTestModel(nil)
	m, _ = press(m, "a") // approve
	m, _ = press(m, "e") // open editor (starts in the body)
	if m.mode != modeEdit {
		t.Fatalf("mode = %v, want modeEdit", m.mode)
	}

	m, _ = press(m, "X") // type a character
	next, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = next.(Model)

	if m.mode != modeList {
		t.Fatalf("mode = %v, want modeList", m.mode)
	}
	if m.jobs[0].Draft.Body != "XBody A" {
		t.Errorf("body = %q, want %q", m.jobs[0].Draft.Body, "XBody A")
	}
	if m.jobs[0].Approved {
		t.Error("expected approval to be reset after an edit")
	}
}

func TestSendRunsOneAtATime(t *testing.T) {
	var sent []string
	send := func(_ context.Context, job domain.EmailJob) error {
		sent = append(sent, job.Company)
		if job.Company == "Globex" {
			return errors.New("boom")
		}
		return nil
	}

	m := newTestModel(send)
	m, _ = press(m, "A")
	m, _ = press(m, "s")
	if m.mode != modeConfirm {
		t.Fatalf("mode = %v, want modeConfirm", m.mode)
	}

	m, cmd := press(m, "y")
	if cmd == nil {
		t.Fatal("expected a send command")
	}
	if m.jobs[0].Status != domain.Sending || m.jobs[1].Status != domain.Sending {
		t.Fatal("expected both jobs marked as sending/queued")
	}
	if len(sent) != 0 {
		t.Fatalf("nothing should be sent until the cmd runs, got %v", sent)
	}

	// Run the first send. Only Acme goes out; the model then asks for the next one.
	next, cmd := m.Update(cmd())
	m = next.(Model)
	if len(sent) != 1 || sent[0] != "Acme" {
		t.Fatalf("sent = %v, want [Acme]", sent)
	}
	if m.jobs[0].Status != domain.Sent {
		t.Errorf("Acme status = %v, want sent", m.jobs[0].Status)
	}
	if cmd == nil {
		t.Fatal("expected the next send command")
	}

	// Run the second send, which fails.
	next, cmd = m.Update(cmd())
	m = next.(Model)
	if cmd != nil {
		t.Error("expected no more commands")
	}
	if m.jobs[1].Status != domain.Failed {
		t.Errorf("Globex status = %v, want failed", m.jobs[1].Status)
	}
	if m.jobs[1].Approved {
		t.Error("a failed job should need re-approval")
	}
}

func TestViewDoesNotPanic(t *testing.T) {
	m := newTestModel(nil)
	for _, key := range []string{"j", "?", "x", "e"} {
		m, _ = press(m, key)
		_ = m.View()
	}
	// Empty model.
	empty := New(nil, nil)
	next, _ := empty.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_ = next.(Model).View()
}

func TestEditorWrapAndMove(t *testing.T) {
	e := newEditor("subj", "aaaaaaaaaa\nbb") // 10 chars, then 2
	e.w, e.h = 4, 5
	e.field = 1

	e.vertical(1) // wrapped row 1 of line 0
	if e.row != 0 || e.col != 4 {
		t.Fatalf("after down: row=%d col=%d, want row=0 col=4", e.row, e.col)
	}

	e.vertical(1) // row 2 of line 0
	e.vertical(1) // next logical line
	if e.row != 1 {
		t.Fatalf("row = %d, want 1", e.row)
	}

	e.backspace()
	e.backspace()
	e.backspace() // joins lines
	if len(e.body) != 1 {
		t.Fatalf("lines = %d, want 1 after join", len(e.body))
	}
}

func TestSendSkipsAlreadyContacted(t *testing.T) {
	send := func(context.Context, domain.EmailJob) error {
		return domain.ErrAlreadyContacted
	}

	m := newTestModel(send)
	m, _ = press(m, "a")
	m, _ = press(m, "s")
	m, cmd := press(m, "y")

	next, _ := m.Update(cmd())
	m = next.(Model)

	if m.jobs[0].Status != domain.Failed {
		t.Errorf("status = %v, want failed", m.jobs[0].Status)
	}
	if m.jobs[0].Approved {
		t.Error("skipped job should not stay approved")
	}
}

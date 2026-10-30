package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	if len(m.Jobs) == 0 {
		return tea.NewView("No emails to review.")
	}

	job := m.Jobs[m.Selected]

	return tea.NewView(fmt.Sprintf(
		"Company: %s\nTo: %s\nSubject: %s\n\n%s\n",
		job.Prospect.Company,
		job.Prospect.Email,
		job.Draft.Subject,
		job.Draft.Body,
	))
}

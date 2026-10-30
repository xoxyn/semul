package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/xoxyn/semul/internal/domain"
)

type Model struct {
	Jobs     []domain.EmailJob
	Selected int
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle keyboard input here.
	return m, nil
}

package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/esferadigital/cadence/internal/timer"
)

type model struct {
	remaining time.Duration
	humanIdx  int
	kind      timer.PhaseKind
	done      bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		m.remaining = msg.PhaseRemaining
		m.humanIdx = msg.PhaseHumanIdx
		m.kind = msg.PhaseKind
		return m, nil
	case timer.TimerFinishedMsg:
		m.done = true
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	var s string
	if m.done {
		s = "Nice job!"
	} else {
		s = fmt.Sprintf("%s %d\n%s", m.kind.Name(), m.humanIdx, formatRemaining(m.remaining))
	}
	return s
}

func NewModel() model {
	return model{}
}

func formatRemaining(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	if totalSeconds < 0 {
		totalSeconds = -totalSeconds
	}
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

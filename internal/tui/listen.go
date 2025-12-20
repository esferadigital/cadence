package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/esferadigital/cadence/internal/timer"
)

func Listen(program *tea.Program, messages <-chan timer.TimerMsg) {
	for message := range messages {
		program.Send(message)
	}
}

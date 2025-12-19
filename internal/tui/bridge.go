package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/esferadigital/cadence/internal/timer"
)

func Bridge(program *tea.Program, messages <-chan timer.TimerMsg) {
	for {
		message, ok := <-messages
		if !ok {
			return
		}
		program.Send(message)
	}
}

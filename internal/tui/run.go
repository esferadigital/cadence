package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/esferadigital/cadence/internal/logs"
	"github.com/esferadigital/cadence/internal/pomodoro"
)

func Run(events <-chan pomodoro.Event, machine *pomodoro.Machine, appLogger logs.Logger) {
	p := tea.NewProgram(newModel(machine, appLogger), tea.WithAltScreen())

	go func() {
		for event := range events {
			p.Send(event)
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

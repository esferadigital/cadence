package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/esferadigital/cadence/internal/bus"
	"github.com/esferadigital/cadence/internal/notifier"
	"github.com/esferadigital/cadence/internal/timer"
	"github.com/esferadigital/cadence/internal/tui"
)

const WorkDuration = 10 * time.Second
const BreakDuration = 5 * time.Second
const WorkPhases = 4
const Interval = 500 * time.Millisecond

func main() {
	p := tea.NewProgram(tui.NewModel())

	t := timer.New(Interval, WorkDuration, BreakDuration, WorkPhases)

	eventBus := bus.New()
	go eventBus.Run(t.Messages())

	go tui.Listen(p, eventBus.Subscribe())
	go notifier.Listen(eventBus.Subscribe())

	t.Start()
	go t.Run()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

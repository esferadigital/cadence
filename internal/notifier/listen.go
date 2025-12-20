package notifier

import (
	"fmt"

	"github.com/esferadigital/cadence/internal/timer"
	"github.com/gen2brain/beeep"
)

const AppName = "Cadence"

func Listen(messages <-chan timer.TimerMsg) {
	for msg := range messages {
		switch msg := msg.(type) {
		case timer.PhaseFinishedMsg:
			var text string
			if msg.PhaseKind == timer.PhaseWork {
				text = "Take 5"
			} else {
				text = "Time to grind"
			}
			title := fmt.Sprintf("%s %d finished", msg.PhaseKind.Name(), msg.PhaseHumanIdx)
			notify(title, text)
		case timer.TimerFinishedMsg:
			notify("Timer finished", "Nice job")
			return
		}
	}
}

func notify(title string, text string) error {
	beeep.AppName = AppName
	return beeep.Notify(title, text, "")
}

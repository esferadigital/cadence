package notifier

import (
	"fmt"

	"github.com/esferadigital/cadence/internal/timer"
	"github.com/gen2brain/beeep"
)

const AppName = "Cadence"

func Listen(events <-chan timer.TimerEvent) {
	for {
		event, ok := <-events
		if !ok {
			return
		}
		switch event.Kind {
		case timer.PhaseFinished:
			title := fmt.Sprintf("Phase %d finished", event.Data.PhaseIdx)
			send(title, "Keep grinding")
		case timer.TimerFinished:
			send("Timer finished", "Nice job")
			return
		}
	}
}

func send(title string, msg string) error {
	beeep.AppName = AppName
	return beeep.Notify(title, msg, "")
}

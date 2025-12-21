package notify

import (
	"fmt"

	"github.com/esferadigital/cadence/internal/pomodoro"
	"github.com/gen2brain/beeep"
)

const AppName = "Cadence"

func Run(events <-chan pomodoro.Event) {
	go func() {
		for event := range events {
			switch event := event.(type) {
			case pomodoro.EventPhaseFinished:
				notifyPhaseFinished(event)
			case pomodoro.EventTimerFinished:
				notifyTimerFinished()
			}
		}
	}()
}

func notifyPhaseFinished(event pomodoro.EventPhaseFinished) {
	var text string
	if event.Phase.Kind == pomodoro.PhaseWork {
		text = "Take 5"
	} else {
		text = "Time to grind"
	}
	title := fmt.Sprintf("%s %d finished", event.Phase.Kind, event.Phase.HumanIdx)
	notify(title, text)
}

func notifyTimerFinished() {
	notify("Timer finished", "Nice job")
}

func notify(title string, text string) error {
	beeep.AppName = AppName
	return beeep.Notify(title, text, "")
}

package main

import (
	"time"

	"github.com/esferadigital/cadence/internal/notifier"
	"github.com/esferadigital/cadence/internal/timer"
)

const WorkDuration = 10 * time.Second
const BreakDuration = 5 * time.Second
const WorkPhases = 4
const Interval = 1 * time.Second

func main() {
	t := timer.New(WorkDuration, BreakDuration, WorkPhases)

	go notifier.Listen(t.Events())

	t.Start()
	for !t.IsFinished() {
		t.Tick()
		time.Sleep(Interval)
	}
}

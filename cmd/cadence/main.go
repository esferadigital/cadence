package main

import (
	"flag"

	"github.com/esferadigital/cadence/internal/logs"
	"github.com/esferadigital/cadence/internal/notify"
	"github.com/esferadigital/cadence/internal/pomodoro"
	"github.com/esferadigital/cadence/internal/tui"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	appLogger := logs.New()
	defer appLogger.Clean()
	appLogger.SetEnabled(*debug)

	m := pomodoro.NewMachine(appLogger)
	m.Run()

	notifySub := m.Subscribe()
	notify.Run(notifySub)

	tuiSub := m.Subscribe()
	tui.Run(tuiSub, m, appLogger)
}

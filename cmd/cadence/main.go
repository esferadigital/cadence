package main

import (
	"flag"

	"github.com/diegoserranor/cadence/internal/config"
	"github.com/diegoserranor/cadence/internal/logs"
	"github.com/diegoserranor/cadence/internal/notify"
	"github.com/diegoserranor/cadence/internal/pomodoro"
	"github.com/diegoserranor/cadence/internal/tui"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	workMinutes := flag.Int("work", 0, "work phase length in minutes")
	breakMinutes := flag.Int("break", 0, "break phase length in minutes")
	flag.Parse()

	appLogger := logs.New()
	defer appLogger.Clean()
	appLogger.SetEnabled(*debug)

	cfg, err := config.LoadWithOverrides(*workMinutes, *breakMinutes)
	if err != nil && appLogger != nil {
		appLogger.Printf("config load failed: %v", err)
	}

	m := pomodoro.NewMachine(appLogger, cfg.WorkMinutes, cfg.BreakMinutes, cfg.WorkPhases)
	m.Run()

	notifySub := m.Subscribe()
	notify.Run(notifySub)

	tuiSub := m.Subscribe()
	tui.Run(tuiSub, m, cfg, appLogger)
}

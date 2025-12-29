package pomodoro

import "time"

// Interface meant to be implemented to apply state transitions.
// See `internal/pomodoro/state.go`.
type processor interface {
	apply(cmd command) transition
	tick() transition
}

type command int

const (
	commandStart command = iota
	commandPause
	commandResume
	commandSkipBreak
	commandGetState
)

type transition struct {
	From        stateSnapshot
	To          stateSnapshot
	Completions []phaseCompletion
	Finished    bool
	EmitState   bool
}

type stateSnapshot struct {
	Phase      PhaseSnapshot
	Status     TimerStatus
	WorkPhases int
}

type phaseCompletion struct {
	Phase PhaseSnapshot
}

type PhaseSnapshot struct {
	Idx       int
	HumanIdx  int
	Kind      PhaseKind
	Duration  time.Duration
	Remaining time.Duration
}

type PhaseDetail struct {
	Kind     PhaseKind
	Duration time.Duration
}

type PhaseKind string

const (
	PhaseWork  PhaseKind = "Work"
	PhaseBreak PhaseKind = "Break"
)

func phaseHumanIdx(phase int) int {
	return phase/2 + 1
}

type TimerStatus int

const (
	StatusInit TimerStatus = iota
	StatusRunning
	StatusPaused
	StatusFinished
)

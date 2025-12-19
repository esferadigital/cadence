package timer

import (
	"time"
)

// ---- status ----

type TimerStatus int

const (
	StatusInit TimerStatus = iota
	StatusRunning
	StatusPaused
	StatusFinished
)

// ---- phases ----

type PhaseKind int

const (
	PhaseWork PhaseKind = iota
	PhaseBreak
)

func (pk PhaseKind) Name() string {
	var name string
	if pk == PhaseWork {
		name = "Work"
	} else {
		name = "Break"
	}
	return name
}

type PhaseDetail struct {
	Kind     PhaseKind
	Duration time.Duration
}

// ---- events ----

type TimerMsg interface{}

type TickMsg struct {
	PhaseRemaining time.Duration
	PhaseIdx       int
	PhaseHumanIdx  int
	PhaseKind      PhaseKind
}

type PhaseFinishedMsg struct {
	PhaseIdx      int
	PhaseHumanIdx int
	PhaseKind     PhaseKind
}

type TimerFinishedMsg struct{}

// ---- timer ----

type Timer struct {
	workDur       time.Duration
	breakDur      time.Duration
	phaseCnt      int
	phaseIdx      int
	phaseElapsed  time.Duration
	phaseLastTick time.Time
	status        TimerStatus
	messages      chan TimerMsg
}

func New(workDur time.Duration, breakDur time.Duration, workPhases int) Timer {
	return Timer{
		workDur:  workDur,
		breakDur: breakDur,
		phaseCnt: (workPhases * 2) - 1,
		phaseIdx: 0,
		status:   StatusInit,
		messages: make(chan TimerMsg),
	}
}

func (t *Timer) Events() <-chan TimerMsg {
	return t.messages
}

func (t *Timer) Start() {
	if t.status == StatusInit {
		t.phaseElapsed = time.Second * 0
		t.phaseLastTick = time.Now()
		t.status = StatusRunning
	}
}

// TODO:
// - Handle case where the timer is not actually running
// - Handle case where the timer's last tick is at it's default value
func (t *Timer) Tick() {
	delta := time.Since(t.phaseLastTick)
	for delta > 0 {
		phase := t.phaseDetail()

		// Handle ticks that occur while we are still in the current phase duration
		if delta+t.phaseElapsed < phase.Duration {
			t.phaseElapsed += delta
			break
		}

		// Handle ticks that occur after (i.e. phase is finished)
		// Check if all the phases are complete
		next := t.phaseIdx + 1
		if next >= t.phaseCnt {
			t.status = StatusFinished
			message := TimerFinishedMsg{}
			t.messages <- message
			close(t.messages)
			return
		}

		// Send a message saying the phase has finished
		message := PhaseFinishedMsg{
			PhaseIdx:      t.phaseIdx,
			PhaseHumanIdx: humanIdx(t.phaseIdx),
			PhaseKind:     phase.Kind,
		}
		t.messages <- message

		// Subtract remaining phase duration from delta
		remaining := t.phaseRemaining()
		delta -= remaining

		// Advance to the next phase
		t.phaseIdx = next
		t.phaseElapsed = time.Duration(0)
	}

	message := TickMsg{
		PhaseRemaining: t.phaseRemaining(),
		PhaseIdx:       t.phaseIdx,
		PhaseHumanIdx:  humanIdx(t.phaseIdx),
		PhaseKind:      t.phaseDetail().Kind,
	}
	t.messages <- message

	t.phaseLastTick = time.Now()
}

func (t *Timer) IsFinished() bool {
	return t.status == StatusFinished
}

func (t *Timer) phaseRemaining() time.Duration {
	phase := t.phaseDetail()
	return phase.Duration - t.phaseElapsed
}

func (t *Timer) phaseDetail() PhaseDetail {
	detail := PhaseDetail{}
	if t.phaseIdx%2 == 0 {
		detail.Kind = PhaseWork
		detail.Duration = t.workDur
	} else {
		detail.Kind = PhaseBreak
		detail.Duration = t.breakDur
	}
	return detail
}

func humanIdx(phase int) int {
	return phase/2 + 1
}

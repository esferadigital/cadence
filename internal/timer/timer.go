package timer

import (
	"log"
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

type PhaseDetail struct {
	Kind     PhaseKind
	Duration time.Duration
}

// ---- events ----

type EventKind int

const (
	Tick EventKind = iota
	PhaseFinished
	TimerFinished
)

// Data sent with event
type EventData struct {
	// Sent for `Tick` and `PhaseFinished`
	PhaseIdx int

	// Sent for `Tick`
	PhaseRemaining time.Duration
}

type TimerEvent struct {
	// Always sent
	Kind EventKind

	// Sent for `Tick` and `PhaseFinished`
	Data EventData
}

// ---- timer ----

type Timer struct {
	workDur       time.Duration
	breakDur      time.Duration
	phaseCnt      int
	phaseIdx      int
	phaseElapsed  time.Duration
	phaseLastTick time.Time
	status        TimerStatus
	events        chan TimerEvent
}

func New(workDur time.Duration, breakDur time.Duration, workPhases int) Timer {
	return Timer{
		workDur:  workDur,
		breakDur: breakDur,
		phaseCnt: (workPhases * 2) - 1,
		phaseIdx: 0,
		status:   StatusInit,
		events:   make(chan TimerEvent),
	}
}

func (t *Timer) Events() <-chan TimerEvent {
	return t.events
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
	log.Println("tick, status:", t.status)
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
			event := TimerEvent{
				Kind: TimerFinished,
			}
			t.events <- event
			close(t.events)
			return
		}

		// Send an event saying the phase has finished
		data := EventData{
			PhaseIdx: t.phaseIdx,
		}
		event := TimerEvent{
			Kind: PhaseFinished,
			Data: data,
		}
		t.events <- event

		// Subtract remaining phase duration from delta
		remaining := t.phaseRemaining()
		delta -= remaining

		// Advance to the next phase
		t.phaseIdx = next
		t.phaseElapsed = time.Duration(0)
	}

	data := EventData{
		PhaseIdx:       t.phaseIdx,
		PhaseRemaining: t.phaseRemaining(),
	}
	event := TimerEvent{
		Kind: Tick,
		Data: data,
	}
	t.events <- event

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

// func humanIdx(phase int) int {
// 	return phase/2 + 1
// }

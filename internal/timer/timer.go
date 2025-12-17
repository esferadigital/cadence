package timer

import (
	"log"
	"time"
)

type TimerStatus int

const (
	StatusInit TimerStatus = iota
	StatusRunning
	StatusPaused
	StatusFinished
)

type PhaseType int

const (
	PhaseWork PhaseType = iota
	PhaseBreak
)

type Timer struct {
	workDur       time.Duration
	breakDur      time.Duration
	phaseCnt      int
	phaseIdx      int
	phaseElapsed  time.Duration
	phaseLastTick time.Time
	status        TimerStatus
	// sender
}

func New(workDur time.Duration, breakDur time.Duration, workPhases int) Timer {
	return Timer{
		workDur:  workDur,
		breakDur: breakDur,
		phaseCnt: (workPhases * 2) - 1,
		phaseIdx: 0,
		status:   StatusInit,
	}
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
		var phaseType PhaseType
		var phaseDur time.Duration
		if t.phaseIdx%2 == 0 {
			phaseType = PhaseWork
			phaseDur = t.workDur
		} else {
			phaseType = PhaseBreak
			phaseDur = t.breakDur
		}

		// Handle ticks that occur while we are still in the current phase duration
		if delta+t.phaseElapsed < phaseDur {
			t.phaseElapsed += delta
			break
		}

		// Handle ticks that occur after (i.e. phase is finished)
		// Subtract remaining phase duration from delta
		delta -= phaseDur - t.phaseElapsed

		finishedPhase := t.phaseIdx/2 + 1
		if phaseType == PhaseWork {
			// Send event that work is finished
			log.Printf("Finished work phase %d", finishedPhase)
		} else {
			// Send event that break is finished
			log.Printf("Finished break phase %d", finishedPhase)
		}

		// Advance to the next phase
		t.phaseIdx += 1
		t.phaseElapsed = time.Duration(0)

		// Check if new phase is actually beyond the expected phases
		if t.phaseIdx >= t.phaseCnt {
			t.status = StatusFinished

			// Send event that the timer is finished
			log.Println("Time is finished; nice job")

			return
		}
	}

	t.phaseLastTick = time.Now()
	// TODO: Send timer tick event
	log.Printf("Tick! %d", t.phaseElapsed)
}

func (t *Timer) IsFinished() bool {
	return t.status == StatusFinished
}

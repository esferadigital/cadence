package pomodoro

import (
	"time"
)

type state struct {
	workDur       time.Duration
	breakDur      time.Duration
	workPhases    int
	phaseCnt      int
	phaseIdx      int
	phaseElapsed  time.Duration
	phaseLastTick time.Time
	phaseLastWall time.Time
	status        TimerStatus
}

type advanceDelta struct {
	completions []phaseCompletion
	finished    bool
}

func newState(workDur time.Duration, breakDur time.Duration, workPhases int) *state {
	return &state{
		workDur:    workDur,
		breakDur:   breakDur,
		workPhases: workPhases,
		phaseCnt:   (workPhases * 2) - 1,
		phaseIdx:   0,
		status:     StatusInit,
	}
}

func (s *state) apply(cmd command) transition {
	before := s.snapshot()
	after := before
	delta := advanceDelta{}
	emitState := false

	switch cmd {
	case commandStart:
		if s.start() {
			emitState = true
		}
	case commandPause:
		var paused bool
		delta, paused = s.pause()
		if paused {
			emitState = true
		}
	case commandResume:
		if s.resume() {
			emitState = true
		}
	case commandGetState:
		emitState = true
	}

	if emitState || delta.finished || len(delta.completions) > 0 {
		after = s.snapshot()
	}

	return transition{
		From:        before,
		To:          after,
		Completions: delta.completions,
		Finished:    delta.finished,
		EmitState:   emitState,
	}
}

func (s *state) tick() transition {
	before := s.snapshot()
	if s.status != StatusRunning {
		return transition{
			From:      before,
			To:        before,
			EmitState: false,
		}
	}

	elapsed := elapsedSinceLastTick(s.phaseLastTick, s.phaseLastWall)
	delta := s.advance(elapsed)
	s.phaseLastTick = time.Now()
	s.phaseLastWall = nowWallClock()

	after := s.snapshot()
	return transition{
		From:        before,
		To:          after,
		Completions: delta.completions,
		Finished:    delta.finished,
		EmitState:   true,
	}
}

func (s *state) start() bool {
	if s.status == StatusInit {
		s.phaseElapsed = time.Second * 0
		s.phaseLastTick = time.Now()
		s.phaseLastWall = nowWallClock()
		s.status = StatusRunning
		return true
	}
	return false
}

func (s *state) pause() (advanceDelta, bool) {
	if s.status == StatusRunning {
		delta := advanceDelta{}
		if !s.phaseLastTick.IsZero() && !s.phaseLastWall.IsZero() {
			delta = s.advance(elapsedSinceLastTick(s.phaseLastTick, s.phaseLastWall))
		}
		if s.status == StatusRunning {
			s.status = StatusPaused
		}
		s.phaseLastTick = time.Now()
		s.phaseLastWall = nowWallClock()
		return delta, true
	}
	return advanceDelta{}, false
}

func (s *state) resume() bool {
	if s.status == StatusPaused {
		s.phaseLastTick = time.Now()
		s.phaseLastWall = nowWallClock()
		s.status = StatusRunning
		return true
	}
	return false
}

func (s *state) advance(elapsed time.Duration) advanceDelta {
	completions := make([]phaseCompletion, 0, 1)
	for elapsed > 0 {
		phase := s.phaseDetail()
		phaseRemaining := phase.Duration - s.phaseElapsed
		if elapsed < phaseRemaining {
			s.phaseElapsed += elapsed
			return advanceDelta{completions: completions, finished: false}
		}

		// From this point forward the phase we were tracking has already ended
		// So:
		// - Subtract the time that remained from that phase
		// - Calculate the next index
		// - Exit early if the next index is beyond the expected phase count (time is done)
		elapsed -= phaseRemaining
		nextIdx := s.phaseIdx + 1
		if nextIdx >= s.phaseCnt {
			s.status = StatusFinished
			s.phaseElapsed = phase.Duration
			return advanceDelta{completions: completions, finished: true}
		}

		// From this point forward we still have phases to complete,
		// but we note that the previous phase has been completed
		completions = append(completions, phaseCompletion{
			Phase: PhaseSnapshot{
				Idx:       s.phaseIdx,
				HumanIdx:  phaseHumanIdx(s.phaseIdx),
				Kind:      phase.Kind,
				Duration:  phase.Duration,
				Remaining: 0,
			},
		})

		// Update the phase index and reset the phase elapsed time to 0
		// NOTE: The `elapsed` value that this loop tracks it not necessarily 0 at this point
		s.phaseIdx = nextIdx
		s.phaseElapsed = time.Duration(0)
	}
	return advanceDelta{completions: completions, finished: s.status == StatusFinished}
}

func (s *state) snapshot() stateSnapshot {
	return stateSnapshot{
		Phase:      s.phaseSnapshot(),
		Status:     s.status,
		WorkPhases: s.workPhases,
	}
}

func (s *state) phaseSnapshot() PhaseSnapshot {
	phase := s.phaseDetail()
	return PhaseSnapshot{
		Idx:       s.phaseIdx,
		HumanIdx:  phaseHumanIdx(s.phaseIdx),
		Kind:      phase.Kind,
		Duration:  phase.Duration,
		Remaining: phase.Duration - s.phaseElapsed,
	}
}

func (s *state) phaseDetail() PhaseDetail {
	detail := PhaseDetail{}
	if s.phaseIdx%2 == 0 {
		detail.Kind = PhaseWork
		detail.Duration = s.workDur
	} else {
		detail.Kind = PhaseBreak
		detail.Duration = s.breakDur
	}
	return detail
}

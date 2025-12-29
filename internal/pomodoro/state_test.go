package pomodoro

import (
	"testing"
	"time"
)

func TestAdvanceSkipsPhasesOnLongElapsed(t *testing.T) {
	s := newState(25*time.Minute, 5*time.Minute, 4)
	if !s.start() {
		t.Fatal("expected start to succeed")
	}

	delta := s.advance(90 * time.Minute)
	if delta.finished {
		t.Fatal("expected timer to keep running after 90 minutes")
	}
	if len(delta.completions) != 6 {
		t.Fatalf("expected 6 phase completions, got %d", len(delta.completions))
	}
	if delta.completions[0].Phase.Idx != 0 || delta.completions[0].Phase.Kind != PhaseWork {
		t.Fatalf("expected first completion to be phase 0 work, got idx=%d kind=%s", delta.completions[0].Phase.Idx, delta.completions[0].Phase.Kind)
	}
	if delta.completions[5].Phase.Idx != 5 || delta.completions[5].Phase.Kind != PhaseBreak {
		t.Fatalf("expected last completion to be phase 5 break, got idx=%d kind=%s", delta.completions[5].Phase.Idx, delta.completions[5].Phase.Kind)
	}
	if s.phaseIdx != 6 {
		t.Fatalf("expected to land on phase 6 after 90 minutes, got %d", s.phaseIdx)
	}
	if s.phaseElapsed != 0 {
		t.Fatalf("expected phase elapsed to reset after phase advance, got %s", s.phaseElapsed)
	}
}

func TestAdvanceFinishesOnVeryLongElapsed(t *testing.T) {
	s := newState(25*time.Minute, 5*time.Minute, 4)
	if !s.start() {
		t.Fatal("expected start to succeed")
	}

	delta := s.advance(200 * time.Minute)
	if !delta.finished {
		t.Fatal("expected timer to finish after a very long elapsed time")
	}
	if s.status != StatusFinished {
		t.Fatalf("expected status finished, got %v", s.status)
	}
	if s.phaseIdx != 6 {
		t.Fatalf("expected to finish on last phase index 6, got %d", s.phaseIdx)
	}
	if s.phaseElapsed != s.workDur {
		t.Fatalf("expected elapsed to equal full work duration, got %s", s.phaseElapsed)
	}
	if len(delta.completions) != 6 {
		t.Fatalf("expected 6 phase completions before finishing, got %d", len(delta.completions))
	}
}

func TestSkipBreakAdvancesToWork(t *testing.T) {
	s := newState(25*time.Minute, 5*time.Minute, 4)
	if !s.start() {
		t.Fatal("expected start to succeed")
	}

	s.advance(s.workDur)
	if s.phaseDetail().Kind != PhaseBreak {
		t.Fatalf("expected to be on break before skip, got %s", s.phaseDetail().Kind)
	}

	delta, skipped := s.skipBreak()
	if !skipped {
		t.Fatal("expected skip to succeed during break")
	}
	if delta.finished {
		t.Fatal("expected skip break to keep timer running")
	}
	if len(delta.completions) != 1 {
		t.Fatalf("expected 1 completion, got %d", len(delta.completions))
	}
	if delta.completions[0].Phase.Kind != PhaseBreak || delta.completions[0].Phase.Idx != 1 {
		t.Fatalf("expected completion for break phase 1, got idx=%d kind=%s", delta.completions[0].Phase.Idx, delta.completions[0].Phase.Kind)
	}
	if s.phaseDetail().Kind != PhaseWork {
		t.Fatalf("expected to land on work after skip, got %s", s.phaseDetail().Kind)
	}
	if s.phaseElapsed != 0 {
		t.Fatalf("expected phase elapsed to reset after skip, got %s", s.phaseElapsed)
	}
}

func TestSkipBreakNoopDuringWork(t *testing.T) {
	s := newState(25*time.Minute, 5*time.Minute, 4)
	if !s.start() {
		t.Fatal("expected start to succeed")
	}

	delta, skipped := s.skipBreak()
	if skipped {
		t.Fatal("expected skip to be a no-op during work")
	}
	if delta.finished || len(delta.completions) > 0 {
		t.Fatal("expected no completions or finished state when skipping during work")
	}
	if s.phaseDetail().Kind != PhaseWork {
		t.Fatalf("expected to remain on work, got %s", s.phaseDetail().Kind)
	}
}

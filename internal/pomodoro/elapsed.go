package pomodoro

import "time"

const (
	// If wall clock advances meaningfully beyond monotonic, assume sleep and use wall time.
	sleepDetectDriftThreshold = 5 * time.Second
)

// Determines elapsed time. It chooses monotonic elapsed time by default.
// It switches to wall clock when drift suggests a sleep.
// Negative wall deltas get clamped, and an optional cap can be applied to avoid huge jumps.
func elapsedSinceLastTick(lastTick time.Time, lastWall time.Time) time.Duration {
	monoElapsed := time.Since(lastTick)
	wallElapsed := max(time.Since(lastWall), 0)

	drift := max(wallElapsed-monoElapsed, 0)
	elapsed := monoElapsed
	if drift > sleepDetectDriftThreshold {
		elapsed = wallElapsed
	}
	return elapsed
}

// Strips monotonic component so elapsed uses wall clock (continues during sleep).
func nowWallClock() time.Time {
	return time.Now().Round(0)
}

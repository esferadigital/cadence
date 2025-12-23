package pomodoro

import (
	"sync"
	"time"

	"github.com/esferadigital/cadence/internal/logs"
)

const (
	interval      = 250 * time.Millisecond
	workDuration  = 25 * time.Minute
	breakDuration = 5 * time.Minute
	workPhases    = 4
)

// Pomodoro state machine that drives the timer, accepts commands, and broadcasts state change events.
type Machine struct {
	cmds        chan command
	mu          sync.Mutex
	subscribers []chan Event
	processor   processor
	logger      logs.Logger
}

// Create a new pomodoro state machine and receive a pointer to it.
// NOTE: This was developed with the assumption that it is only called once in the application.
// Pass a logger to capture dropped events when debugging.
func NewMachine(appLogger logs.Logger) *Machine {
	m := Machine{
		cmds:        make(chan command, 10),
		subscribers: make([]chan Event, 0),
		processor:   newState(workDuration, breakDuration, workPhases),
		logger:      appLogger,
	}

	return &m
}

// Runs the internal loop that drives the timer in a goroutine.
func (m *Machine) Run() {
	go m.run()
}

// Create and add a unique channel to the machine's subscriptions list.
// Use this channel to receive machine events.
func (m *Machine) Subscribe() <-chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan Event, 10)
	m.subscribers = append(m.subscribers, ch)
	return ch
}

// Starts the timer.
// Only works if status is `StatusInit`, otherwise it is a no-op.
func (m *Machine) Start() {
	m.cmds <- commandStart
}

// Pauses the timer.
// Only works if status is `StatusRunning`, otherwise it is a no-op.
func (m *Machine) Pause() {
	m.cmds <- commandPause
}

// Resumes the timer.
// Only works is status is `StatusPaused`, otherwise it is a no-op.
func (m *Machine) Resume() {
	m.cmds <- commandResume
}

// Requests a snapshot of the current machine state.
// The state is broadcasted with the event `EventStateChanged`.
func (m *Machine) GetState() {
	m.cmds <- commandGetState
}

// Internal loop to run the state machine.
// It forwards commands to the state processor and responds with events on state transitions.
func (m *Machine) run() {
	var ticker *time.Ticker
	var tickCh <-chan time.Time
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()

	for {
		select {
		case cmd := <-m.cmds:
			transition := m.processor.apply(cmd)
			events := eventsFromTransition(transition)
			for _, event := range events {
				m.broadcast(event)
			}

			// Create a new ticker on "start" and "resume".
			// Stop and nil the ticker on "pause".
			switch cmd {
			case commandStart:
				if transition.To.Status == StatusRunning && ticker == nil {
					ticker = time.NewTicker(interval)
					tickCh = ticker.C
				}
			case commandResume:
				if transition.To.Status == StatusRunning && ticker == nil {
					ticker = time.NewTicker(interval)
					tickCh = ticker.C
				}
			case commandPause:
				if ticker != nil {
					ticker.Stop()
					ticker = nil
					tickCh = nil
				}
			case commandGetState:
				// No ticker changes.
			}

		// Advance the timer on every tick.
		case <-tickCh:
			transition := m.processor.tick()
			events := eventsFromTransition(transition)
			for _, event := range events {
				m.broadcast(event)
			}
			if transition.Finished {
				return
			}
		}
	}
}

func (m *Machine) broadcast(event Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ch := range m.subscribers {
		select {
		case ch <- event:
			// Sent
		default:
			// Subscriber is too slow; drop the event to avoid blocking.
			if m.logger != nil {
				m.logger.Printf("dropped event: %T %+v", event, event)
			}
		}
	}
}

// ---- events ----

type Event interface{}

type EventStateChanged struct {
	Phase      PhaseSnapshot
	Status     TimerStatus
	WorkPhases int
}

type EventPhaseFinished struct {
	Phase PhaseSnapshot
}

type EventTimerFinished struct{}

// Build out a list of events from phase transition data.
// Multiple completions could have occurred because of system sleep or idle times.
// It is also possible for the timer to have finished a phase and to want to emit new state.
func eventsFromTransition(transition transition) []Event {
	events := make([]Event, 0, len(transition.Completions)+2)
	for _, completion := range transition.Completions {
		events = append(events, EventPhaseFinished{
			Phase: completion.Phase,
		})
	}
	if transition.Finished {
		events = append(events, EventTimerFinished{})
	}
	if transition.EmitState {
		events = append(events, stateChangedEvent(transition.To))
	}
	return events
}

// Build out the state changed event with a snapshot of the timer state.
func stateChangedEvent(snapshot stateSnapshot) EventStateChanged {
	return EventStateChanged{
		Phase:      snapshot.Phase,
		Status:     snapshot.Status,
		WorkPhases: snapshot.WorkPhases,
	}
}

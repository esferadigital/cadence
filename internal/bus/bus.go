package bus

import "github.com/esferadigital/cadence/internal/timer"

type Bus struct {
	subs []chan timer.TimerMsg
}

func New() *Bus {
	return &Bus{}
}

func (b *Bus) Subscribe() <-chan timer.TimerMsg {
	ch := make(chan timer.TimerMsg, 1)
	b.subs = append(b.subs, ch)
	return ch
}

func (b *Bus) Run(source <-chan timer.TimerMsg) {
	for msg := range source {
		for _, ch := range b.subs {
			ch <- msg
		}
	}
	for _, ch := range b.subs {
		close(ch)
	}
}

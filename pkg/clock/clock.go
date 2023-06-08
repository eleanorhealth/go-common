package clock

import "time"

type defaultClock struct{}

func NewDefaultClock() Clocker {
	return &defaultClock{}
}

func (t *defaultClock) Now() time.Time {
	return time.Now()
}

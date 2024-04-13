package clock

import "time"

type Clocker interface {
	Now() time.Time
}

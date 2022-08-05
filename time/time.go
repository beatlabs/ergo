package time

import "time"

type Time struct{}

func (w Time) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func (w Time) Now() time.Time {
	return time.Now()
}

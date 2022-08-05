package mock

import "time"

// Time is a mock implementation.
type Time struct {
	CurrentTime time.Time
}

// NewMockedTime creates a new mocked time implementation.
func NewMockedTime(initialTime time.Time) *Time {
	return &Time{CurrentTime: initialTime}
}

// Sleep mocks the sleep action, adding the duration to the time.
func (t Time) Sleep(duration time.Duration) {
	t.CurrentTime = t.CurrentTime.Add(duration)
}

// Now returns the mocked time.
func (t Time) Now() time.Time {
	return t.CurrentTime
}

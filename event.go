package invasion

import "fmt"

// NewEvent creates regular non-important event.
// FIXME 30% of cpu time is spent in fmt.Sprintf and 40% of whole memory allocated per simulation.
// can be changed to lazy evaluation
func NewEvent(format string, a ...interface{}) Event {
	return Event{Data: fmt.Sprintf(format, a...)}
}

// NewImportantEvent creates important events.
func NewImportantEvent(format string, a ...interface{}) Event {
	ev := NewEvent(format, a...)
	ev.Important = true
	return ev
}

// Event can be used to track progress of the invasion.
type Event struct {
	Important bool
	Data      string
}

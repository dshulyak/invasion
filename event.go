package invasion

import "fmt"

// NewEvent creates regular non-important event.
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

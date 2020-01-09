package invasion

import "fmt"

func NewEvent(format string, a ...interface{}) Event {
	return Event{Data: fmt.Sprintf(format, a...)}
}

func NewImportantEvent(format string, a ...interface{}) Event {
	ev := NewEvent(format, a...)
	ev.Important = true
	return ev
}

type Event struct {
	Important bool
	Data      string
}

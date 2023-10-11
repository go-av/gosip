package message

func NewEventHeader(event string) *EventHeader {
	ua := EventHeader(event)
	return &ua
}

type EventHeader string

func (event *EventHeader) Name() string {
	return "Event"
}

func (event EventHeader) Value() string {
	return string(event)
}

func (event *EventHeader) Clone() Header {
	return event
}

func init() {
	defaultHeaderParsers.Register(NewEventHeader(""))
}

func (EventHeader) Parse(data string) (Header, error) {
	return NewEventHeader(data), nil
}

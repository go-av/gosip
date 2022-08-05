package message

func NewAllowEventHeader(event string) *AllowEventHeader {
	id := AllowEventHeader(event)
	return &id
}

type AllowEventHeader string

func (event *AllowEventHeader) Name() string {
	return "Allow-Events"
}

func (event AllowEventHeader) Value() string {
	return string(event)
}

func (event *AllowEventHeader) Clone() Header {
	return event
}

func init() {
	defaultHeaderParsers.Register(NewAllowEventHeader(""))
}

func (AllowEventHeader) Parse(data string) (Header, error) {
	return NewAllowEventHeader(data), nil
}

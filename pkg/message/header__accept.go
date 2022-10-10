package message

func NewAcceptHeader(data string) *AcceptHeader {
	ua := AcceptHeader(data)
	return &ua
}

type AcceptHeader string

func (ua *AcceptHeader) Name() string {
	return "Accept"
}

func (ua AcceptHeader) Value() string {
	return string(ua)
}

func (ua *AcceptHeader) Clone() Header {
	return ua
}

func init() {
	defaultHeaderParsers.Register(NewAcceptHeader(""))
}

func (AcceptHeader) Parse(data string) (Header, error) {
	return NewAcceptHeader(data), nil
}

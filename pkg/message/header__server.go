package message

func NewServerHeader(server string) *ServerHeader {
	ua := ServerHeader(server)
	return &ua
}

type ServerHeader string

func (ua *ServerHeader) Name() string {
	return "Server"
}

func (ua ServerHeader) Value() string {
	return string(ua)
}

func (ua *ServerHeader) Clone() Header {
	return ua
}

func init() {
	defaultHeaderParsers.Register(NewServerHeader(""))
}

func (ServerHeader) Parse(data string) (Header, error) {
	return NewServerHeader(data), nil
}

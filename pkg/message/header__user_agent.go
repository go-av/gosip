package message

func SetUserAgent(ua string) {
	userAgent = ua
}

var userAgent = "go-sip"

func NewUserAgentHeader(userAgent string) *UserAgentHeader {
	ua := UserAgentHeader(userAgent)
	return &ua
}

type UserAgentHeader string

func (ua *UserAgentHeader) Name() string {
	return "User-Agent"
}

func (ua UserAgentHeader) Value() string {
	return string(ua)
}

func (ua *UserAgentHeader) Clone() Header {
	return ua
}

func init() {
	defaultHeaderParsers.Register(NewUserAgentHeader(""))
}

func (UserAgentHeader) Parse(data string) (Header, error) {
	return NewUserAgentHeader(data), nil
}

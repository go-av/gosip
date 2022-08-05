package message

func NewCallIDHeader(callID string) *CallIDHeader {
	id := CallIDHeader(callID)
	return &id
}

type CallIDHeader string

func (callId *CallIDHeader) Name() string {
	return "Call-ID"
}

func (callId CallIDHeader) Value() string {
	return string(callId)
}

func (callId *CallIDHeader) Clone() Header {
	return callId
}

func init() {
	defaultHeaderParsers.Register(NewCallIDHeader(""))
}

func (CallIDHeader) Parse(data string) (Header, error) {
	return NewCallIDHeader(data), nil
}

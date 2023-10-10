package message

func NewReasonHeader(reason string) *ReasonHeader {
	ua := ReasonHeader(reason)
	return &ua
}

type ReasonHeader string

func (reason *ReasonHeader) Name() string {
	return "Reason"
}

func (reason ReasonHeader) Value() string {
	return string(reason)
}

func (reason *ReasonHeader) Clone() Header {
	return reason
}

func init() {
	defaultHeaderParsers.Register(NewReasonHeader(""))
}

func (ReasonHeader) Parse(data string) (Header, error) {
	return NewReasonHeader(data), nil
}

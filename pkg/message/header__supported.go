package message

import "strings"

func NewSupportedHeader(supported []string) *SupportedHeader {
	ua := SupportedHeader(supported)
	return &ua
}

type SupportedHeader []string

func (ua *SupportedHeader) Name() string {
	return "Supported"
}

func (s SupportedHeader) Value() string {
	return strings.Join(s, ", ")
}

func (ua *SupportedHeader) Clone() Header {
	return ua
}

func init() {
	defaultHeaderParsers.Register(NewSupportedHeader(nil))
}

func (SupportedHeader) Parse(data string) (Header, error) {
	return NewSupportedHeader(strings.Split(data, ", ")), nil
}

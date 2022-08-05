package message

import (
	"strings"

	"github.com/go-av/gosip/pkg/method"
)

func NewAllowHeader() AllowHeader {
	return AllowHeader{
		method.REGISTER,
		method.INVITE,
		method.ACK,
		method.BYE,
		method.CANCEL,
		method.UPDATE,
		method.REFER,
		method.PRACK,
		method.SUBSCRIBE,
		method.NOTIFY,
		method.PUBLISH,
		method.MESSAGE,
		method.INFO,
		method.OPTIONS,
	}
}

type AllowHeader []method.Method

func (allow AllowHeader) Name() string {
	return "Allow"
}

func (allow AllowHeader) Value() string {
	parts := make([]string, 0)
	for _, method := range allow {
		parts = append(parts, string(method))
	}
	return strings.Join(parts, ", ")
}

func (allow AllowHeader) Clone() Header {
	if allow == nil {
		var newAllow AllowHeader
		return newAllow
	}

	newAllow := make(AllowHeader, len(allow))
	copy(newAllow, allow)

	return newAllow
}

func init() {
	defaultHeaderParsers.Register(AllowHeader{})
}

func (AllowHeader) Parse(data string) (Header, error) {
	list := strings.Split(data, ",")
	h := AllowHeader{}
	for _, md := range list {
		h = append(h, method.Method(strings.TrimSpace(md)))
	}
	return h, nil
}

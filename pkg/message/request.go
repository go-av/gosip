package message

import (
	"bytes"
	"fmt"

	"github.com/go-av/gosip/pkg/method"
)

type Request interface {
	Message
	Method() method.Method
	Recipient() *Address
	SetRequestFrom(protocol string, address string)
	RequestFrom() (protocol string, address string)
}

func NewRequestMessage(protocol string, method method.Method, recipient *Address) Message {
	req := new(request)
	req.headers = &headers{
		headers: map[string][]Header{},
	}
	req.protocol = protocol
	req.method = method
	req.recipient = recipient
	req.message.startLine = req.StartLine
	// req.body = body

	req.SetHeader(NewUserAgentHeader(userAgent))
	return req
}

type request struct {
	message
	method    method.Method
	recipient *Address
	protocol  string
	address   string
}

func (req *request) StartLine() string {
	buffer := bytes.NewBuffer(nil)

	buffer.WriteString(
		fmt.Sprintf(
			"%s %s %s",
			string(req.Method()),
			req.Recipient(),
			"SIP/2.0",
		),
	)

	return buffer.String()
}

func (req *request) Method() method.Method {
	return req.method
}

func (req *request) Recipient() *Address {
	return req.recipient
}

func (req *request) SetRequestFrom(protocol string, address string) {
	req.protocol = protocol
	req.address = address
}

func (req *request) RequestFrom() (string, string) {
	return req.protocol, req.address
}

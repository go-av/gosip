package message

import (
	"bytes"
	"strings"
	"sync"
)

type MessageBody interface {
	Body() string
	ContentType() string
}

type Message interface {
	StartLine() string
	String() string
	SetBody(body MessageBody)
	Body() MessageBody
	SetHeader(header Header)
	AppendHeader(headers ...Header)
	DelHeader(name string)
	GetHeaders(name string) []Header
	Headers() []Header
	CloneHeader() []Header

	CallID() (*CallIDHeader, bool)
	Via() ([]*ViaHeader, bool)
	From() (*FromHeader, bool)
	To() (*ToHeader, bool)
	CSeq() (*CSeqHeader, bool)
	ContentLength() (*ContentLengthHeader, bool)
	ContentType() (*ContentTypeHeader, bool)
	Contact() (*ContactHeader, bool)
	SetSrc([]byte)
	Src() []byte
}

type message struct {
	*headers
	mu        sync.RWMutex
	body      MessageBody
	startLine func() string
	src       []byte
}

func (msg *message) StartLine() string {
	if msg.startLine == nil {
		return ""
	}
	return msg.startLine()
}

func (msg *message) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(msg.StartLine() + "\r\n")
	// Write headers.
	msg.mu.RLock()
	buf.WriteString(msg.headers.String())
	msg.mu.RUnlock()

	if msg.body == nil {
		buf.WriteString("Content-Length: " + NewContentLengthHeader(0).Value() + "\r\n")
		buf.WriteString("\r\n")
	} else {
		body := msg.body.Body()
		buf.WriteString("Content-Length: " + NewContentLengthHeader(len(body)).Value() + "\r\n")
		buf.WriteString("Content-Type: " + NewContentTypeHeader(msg.body.ContentType()).Value() + "\r\n")
		buf.WriteString("\r\n" + body)
	}

	return buf.String()
}

func (msg *message) Body() MessageBody {
	msg.mu.RLock()
	defer msg.mu.RUnlock()
	return msg.body
}

func (msg *message) SetBody(body MessageBody) {
	msg.mu.RLock()
	defer msg.mu.RUnlock()
	msg.body = body
}

func (msg *message) SetSrc(src []byte) {
	msg.src = src
}

func (msg *message) Src() []byte {
	return msg.src
}

func CopyHeaders(from Message, to Message, names ...string) {
	for _, name := range names {
		name = strings.ToLower(name)
		h := from.GetHeaders(name)
		to.AppendHeader(h...)
	}
}

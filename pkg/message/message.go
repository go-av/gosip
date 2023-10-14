package message

import (
	"bytes"
	"strings"
	"sync"
)

type Headers interface {
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
	WWWAuthenticate() (*WWWAuthenticateHeader, bool)
	Authorization() (*AuthorizationHeader, bool)
	Expires() (*ExpiresHeader, bool)
}

type Message interface {
	SetStartLine(startLine func() string)
	StartLine() string
	String() string
	Body() []byte
	SetBody(contentType string, body []byte)
	Headers
	SetSrc([]byte)
	Src() []byte
}

type message struct {
	*headers
	mu        sync.RWMutex
	startLine func() string
	src       []byte

	body        []byte
	contentType string
}

func (msg *message) SetStartLine(startLine func() string) {
	msg.startLine = startLine
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
		buf.WriteString("Content-Length: " + NewContentLengthHeader(len(msg.body)).Value() + "\r\n")
		buf.WriteString("Content-Type: " + NewContentTypeHeader(msg.contentType).Value() + "\r\n")
		buf.WriteString("\r\n")
		buf.Write(msg.body)
	}

	return buf.String()
}

func (msg *message) Body() []byte {
	return msg.body
}
func (msg *message) SetBody(contentType string, body []byte) {
	msg.contentType = contentType
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

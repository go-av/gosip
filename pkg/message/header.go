package message

import (
	"bytes"
	"strings"
	"sync"
)

var defaultHeaderParsers = &headerParsers{
	headers: make(map[string]HeaderParser),
}

type HeaderParser interface {
	Name() string
	Parse(string) (Header, error)
}

type headerParsers struct {
	mu      sync.RWMutex
	headers map[string]HeaderParser
}

func (p *headerParsers) Register(header HeaderParser) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	p.headers[strings.ToLower(header.Name())] = header
}

func (p *headerParsers) Get(name string) (HeaderParser, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	parser, ok := p.headers[strings.ToLower(name)]
	return parser, ok
}

type Header interface {
	Name() string
	Value() string
	Clone() Header
}

type headers struct {
	mu      sync.RWMutex
	headers map[string][]Header
}

func NewHeaders(hdrs []Header) *headers {
	hs := new(headers)
	hs.headers = make(map[string][]Header)
	hs.AppendHeader(hdrs...)
	return hs
}

func (hs *headers) AppendHeader(headers ...Header) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	for _, header := range headers {
		name := strings.ToLower(header.Name())
		if _, ok := hs.headers[name]; ok {
			hs.headers[name] = append(hs.headers[name], header)
		} else {
			hs.headers[name] = []Header{header}
		}
	}
}

func (hs *headers) SetHeader(header Header) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	hs.headers[strings.ToLower(header.Name())] = []Header{header}
}

func (hs *headers) DelHeader(name string) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	delete(hs.headers, strings.ToLower(name))
}

func (hs *headers) GetHeaders(name string) []Header {
	name = strings.ToLower(name)
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	if hs.headers == nil {
		hs.headers = map[string][]Header{}
	}
	if headers, ok := hs.headers[name]; ok {
		return headers
	}

	return nil
}

func (hs *headers) Headers() []Header {
	hdrs := make([]Header, 0)
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	for _, header := range hs.headers {
		hdrs = append(hdrs, header...)
	}
	return hdrs
}

func (hs *headers) CloneHeader() []Header {
	hdrs := make([]Header, 0)
	for _, header := range hs.Headers() {
		hdrs = append(hdrs, header.Clone())
	}
	return hdrs
}

func (hs *headers) String() string {
	buf := bytes.Buffer{}
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	for _, hds := range hs.headers {
		for _, header := range hds {
			buf.WriteString(header.Name() + ": " + header.Value())
			buf.WriteString("\r\n")
		}
	}
	return buf.String()
}

func (hs *headers) CallID() (*CallIDHeader, bool) {
	vals := hs.GetHeaders("Call-ID")
	if vals == nil {
		return nil, false
	}

	callId, ok := vals[0].(*CallIDHeader)
	if !ok {
		return nil, false
	}
	return callId, true
}

func (hs *headers) Via() ([]*ViaHeader, bool) {
	vals := hs.GetHeaders("Via")
	if vals == nil {
		return nil, false
	}
	var headers []*ViaHeader
	for _, val := range vals {
		via, ok := (val).(*ViaHeader)
		if !ok {
			continue
		}
		headers = append(headers, via)
	}

	return headers, len(headers) > 0
}

func (hs *headers) From() (*FromHeader, bool) {
	vals := hs.GetHeaders("From")
	if vals == nil {
		return nil, false
	}
	from, ok := vals[0].(*FromHeader)
	if !ok {
		return nil, false
	}
	return from, true
}

func (hs *headers) To() (*ToHeader, bool) {
	vals := hs.GetHeaders("To")
	if vals == nil {
		return nil, false
	}
	to, ok := vals[0].(*ToHeader)
	if !ok {
		return nil, false
	}
	return to, true
}

func (hs *headers) CSeq() (*CSeqHeader, bool) {
	vals := hs.GetHeaders("CSeq")
	if vals == nil {
		return nil, false
	}
	cseq, ok := vals[0].(*CSeqHeader)
	if !ok {
		return nil, false
	}
	return cseq, true
}

func (hs *headers) WWWAuthenticate() (*WWWAuthenticateHeader, bool) {
	vals := hs.GetHeaders("WWW-Authenticate")
	if vals == nil {
		return nil, false
	}
	auth, ok := vals[0].(*WWWAuthenticateHeader)
	if !ok {
		return nil, false
	}
	return auth, true
}

func (hs *headers) Authorization() (*AuthorizationHeader, bool) {
	vals := hs.GetHeaders("Authorization")
	if vals == nil {
		return nil, false
	}
	auth, ok := vals[0].(*AuthorizationHeader)
	if !ok {
		return nil, false
	}
	return auth, true
}

func (hs *headers) ContentLength() (*ContentLengthHeader, bool) {
	vals := hs.GetHeaders("Content-Length")
	if vals == nil {
		return nil, false
	}
	contentLength, ok := vals[0].(*ContentLengthHeader)
	if !ok {
		return nil, false
	}
	return contentLength, true
}

func (hs *headers) ContentType() (*ContentTypeHeader, bool) {
	vals := hs.GetHeaders("Content-Type")
	if vals == nil {
		return nil, false
	}
	contentType, ok := vals[0].(*ContentTypeHeader)
	if !ok {
		return nil, false
	}
	return contentType, true
}

func (hs *headers) Contact() (*ContactHeader, bool) {
	vals := hs.GetHeaders("Contact")
	if vals == nil {
		return nil, false
	}
	contactHeader, ok := vals[0].(*ContactHeader)
	if !ok {
		return nil, false
	}
	return contactHeader, true
}

func (hs *headers) Expires() (*ExpiresHeader, bool) {
	vals := hs.GetHeaders("Expires")
	if vals == nil {
		return nil, false
	}
	expiresHeader, ok := vals[0].(*ExpiresHeader)
	if !ok {
		return nil, false
	}
	return expiresHeader, true
}

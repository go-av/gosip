package message

import (
	"fmt"
	"strconv"
)

func NewContentLengthHeader(size int) *ContentLengthHeader {
	l := ContentLengthHeader(size)
	return &l
}

type ContentLengthHeader int

func (contentLength *ContentLengthHeader) Name() string {
	return "Content-Length"
}

func (contentLength ContentLengthHeader) Value() string {
	return fmt.Sprintf("%d", contentLength)
}

func (contentLength *ContentLengthHeader) Clone() Header {
	return contentLength
}

func init() {
	defaultHeaderParsers.Register(NewContentLengthHeader(0))
}

func (ContentLengthHeader) Parse(data string) (Header, error) {
	len, _ := strconv.ParseInt(data, 10, 64)
	return NewContentLengthHeader(int(len)), nil
}

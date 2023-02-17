package message

import (
	"fmt"
	"strconv"
)

func NewExpiresHeader(expires int) *ExpiresHeader {
	l := ExpiresHeader(expires)
	return &l
}

type ExpiresHeader int64

func (expires *ExpiresHeader) Name() string {
	return "Expires"
}

func (expires ExpiresHeader) Value() string {
	return fmt.Sprintf("%d", expires)
}

func (expires *ExpiresHeader) Clone() Header {
	return expires
}

func init() {
	defaultHeaderParsers.Register(NewExpiresHeader(0))
}

func (ExpiresHeader) Parse(data string) (Header, error) {
	expires, _ := strconv.ParseInt(data, 10, 64)
	return NewExpiresHeader(int(expires)), nil
}

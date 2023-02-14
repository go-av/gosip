package utils

import (
	"bytes"
	"strconv"
	"strings"
)

func ParseHostAndPort(text string) (*HostAndPort, error) {
	var host = ""
	var port uint16 = 0
	if n := strings.Index(text, ":"); n > 0 {
		host = text[:n]
		p, _ := strconv.ParseUint(text[n+1:], 10, 64)
		port = uint16(p)
	} else {
		host = text
	}

	return &HostAndPort{
		Host: host,
		Port: port,
	}, nil
}

type HostAndPort struct {
	Host string
	Port uint16
}

func (h *HostAndPort) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(h.Host)
	if h.Port > 0 {
		buf.WriteString(":" + strconv.FormatUint(uint64(h.Port), 10))
	}
	return buf.String()
}

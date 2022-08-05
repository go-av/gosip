package message

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-av/gosip/pkg/types"
)

func NewViaHeader(transport string, host string, port types.Port, params *Params) *ViaHeader {
	return &ViaHeader{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       transport,
		Host:            host,
		Port:            port,
		Params:          params,
	}
}

type ViaHeader struct {
	ProtocolName    string // SIP
	ProtocolVersion string // E.g. '2.0'.
	Transport       string
	Host            string
	Port            types.Port
	Params          *Params
}

func (via *ViaHeader) Name() string {
	return "Via"
}

func (via *ViaHeader) Value() string {
	var buffer bytes.Buffer
	buffer.WriteString(
		fmt.Sprintf(
			"%s/%s/%s %s",
			via.ProtocolName,
			via.ProtocolVersion,
			via.Transport,
			via.Host,
		),
	)
	if via.Port > 0 {
		buffer.WriteString(":" + via.Port.String())
	}

	if via.Params != nil && via.Params.Length() > 0 {
		buffer.WriteString(";")
		buffer.WriteString(via.Params.ToString(";"))
	}

	return buffer.String()
}

func (via *ViaHeader) Clone() Header {
	var newVia *ViaHeader

	if via == nil {
		return newVia
	}

	newVia = &ViaHeader{
		ProtocolName:    via.ProtocolName,
		ProtocolVersion: via.ProtocolVersion,
		Transport:       via.Transport,
		Host:            via.Host,
		Port:            via.Port,
		Params:          via.Params,
	}

	return newVia
}

func init() {
	defaultHeaderParsers.Register(&ViaHeader{})
}

func (ViaHeader) Parse(data string) (Header, error) {
	via := &ViaHeader{}
	via.Params = NewParams()

	parts := strings.Split(data, " ")
	if len(parts) < 2 {
		return nil, errors.New("via data is illegal" + string(data))
	}

	protocols := strings.Split(parts[0], "/")
	if len(protocols) < 3 {
		return nil, errors.New("via SIP Protocol data is illegal:" + parts[0])
	}
	via.ProtocolName = strings.TrimSpace(protocols[0])
	via.ProtocolVersion = strings.TrimSpace(protocols[1])
	via.Transport = strings.TrimSpace(protocols[2])

	viabodys := strings.Split(parts[1], ";")
	for i, body := range viabodys {
		if i == 0 {
			host, port, err := parseHostAndPort(body)
			if err == nil {
				via.Host = host
				via.Port = port
			}
			continue
		}
		param := strings.Split(body, "=")
		if len(param) == 1 {
			via.Params.Set(body, "")
		} else {
			via.Params.Set(param[0], param[1])
		}
	}

	return via, nil
}

func parseHostAndPort(rawText string) (string, types.Port, error) {
	var (
		host    string
		portStr string
		port    types.Port
	)
	if i := strings.Index(rawText, ":"); i == -1 {
		host = rawText
	} else {
		host = rawText[:i]
		portStr = rawText[i+1:]
	}

	if portStr != "" {
		p, _ := strconv.ParseUint(portStr, 10, 64)
		port = types.Port(p)
	}
	return host, port, nil
}

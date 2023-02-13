package message

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func NewViaHeader(transport string, addr string, port uint16, params *Params) *ViaHeader {
	return &ViaHeader{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       strings.ToUpper(transport),
		Addr:            addr,
		Port:            port,
		Params:          params,
	}
}

type ViaHeader struct {
	ProtocolName    string // SIP
	ProtocolVersion string // E.g. '2.0'.
	Transport       string
	Addr            string
	Port            uint16
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
			via.Addr,
		),
	)
	if via.Port > 0 {
		buffer.WriteString(":" + strconv.FormatUint(uint64(via.Port), 10))
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
		Addr:            via.Addr,
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
			addr, port, err := parseAddrAndPort(body)
			if err == nil {
				via.Addr = addr
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

func parseAddrAndPort(rawText string) (string, uint16, error) {
	var (
		host    string
		portStr string
		port    uint16
	)
	if i := strings.Index(rawText, ":"); i == -1 {
		host = rawText
	} else {
		host = rawText[:i]
		portStr = rawText[i+1:]
	}

	if portStr != "" {
		p, _ := strconv.ParseUint(portStr, 10, 64)
		port = uint16(p)
	}
	return host, port, nil
}

package gb28181

import (
	"encoding/xml"
	"sync"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type MessageReceive struct {
	CmdType CmdType `xml:"CmdType"`
	SN      int     `xml:"SN"`
}

type Query struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int64    `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

func NewGB28181(server server.Server, handler GB28181Handler) *GB28181 {
	return &GB28181{
		server:  server,
		handler: handler,
		cache:   &sync.Map{},
	}
}

type GB28181 struct {
	server  server.Server
	handler GB28181Handler
	cache   *sync.Map
}

func (g *GB28181) Handler(body []byte) (*server.Response, error) {
	if body == nil {
		return nil, nil
	}

	message := &MessageReceive{}
	if err := utils.XMLDecode(body, message); err != nil {
		return nil, err
	}

	switch message.CmdType {
	case CmdType__Catalog:
		return g.Catalog(body)
	case CmdType__Keepalive:
		return g.Keepalive(body)
	case CmdType__DeviceInfo:
		return g.DeviceInfo(body)
	}
	return server.NewResponse(200, "success."), nil
}

func (g *GB28181) SendMessage(client server.Client, msg any) (message.Body, error) {
	data, err := xml.MarshalIndent(msg, " ", "")
	if err != nil {
		return nil, err
	}
	return g.server.SendMessage(client, message.NewBody(string(message.ContentType__MANSCDP_XML), data))
}

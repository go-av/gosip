package gb28181

import (
	"encoding/xml"
	"sync"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type MessageReceive struct {
	CmdType CmdType `xml:"CmdType"`
	SN      int     `xml:"SN"`
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
	case CmdType__DeviceStatus:
		return g.DeviceStatus(body)
	case CmdType__PresetQuery:
		return g.PresetQuery(body)
	case CmdType__ConfigDownload:
		return g.ConfigDownload(body)
	}

	return server.NewResponse(200, "success."), nil
}

func (g *GB28181) SendMessage(client server.Client, data any) (message.Body, error) {
	content, err := xml.MarshalIndent(data, " ", "")
	if err != nil {
		return nil, err
	}

	protocol, address := client.Transport()

	hostAndPort, _ := utils.ParseHostAndPort(address)
	clientAddress := message.NewAddress(client.User(), hostAndPort.Host, hostAndPort.Port).WithDomain(g.handler.Realm())
	msg := message.NewRequestMessage(protocol, method.MESSAGE, clientAddress)
	msg.AppendHeader(
		message.NewViaHeader(protocol, g.server.ServerAddress().Host, g.server.ServerAddress().Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewCSeqHeader(1, method.MESSAGE),
		message.NewFromHeader("", g.server.ServerAddress().Clone().SetUser(g.handler.ServerSIPID()).WithDomain(g.handler.Realm()), message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader("", clientAddress, nil),
		message.NewMaxForwardsHeader(70),
	)
	msg.SetBody(string(message.ContentType__MANSCDP_XML), append([]byte("<?xml version=\"1.0\" encoding=\"GBK\"?>\r\n"), content...))
	return g.server.SendMessage(client, msg.(message.Request))
}

package gb28181

import (
	"context"
	"encoding/xml"
	"strings"

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
	}
}

type GB28181 struct {
	server  server.Server
	handler GB28181Handler
}

func (g *GB28181) Handler(ctx context.Context, client server.Client, body []byte) (*server.Response, error) {
	if body == nil {
		return nil, nil
	}

	message := &MessageReceive{}
	if err := utils.XMLDecode(body, message); err != nil {
		return nil, err
	}

	switch message.CmdType {
	case CmdType__Catalog:
		return g.Catalog(ctx, client, body)
	case CmdType__Keepalive:
		return g.Keepalive(ctx, client, body)
	case CmdType__DeviceInfo:
		return g.DeviceInfo(ctx, client, body)
	case CmdType__DeviceStatus:
		return g.DeviceStatus(ctx, client, body)
	case CmdType__PresetQuery:
		return g.PresetQuery(ctx, client, body)
	case CmdType__ConfigDownload:
		return g.ConfigDownload(ctx, client, body)
	case CmdType__Broadcast:
		return g.Broadcast(ctx, client, body)
	}

	return server.NewResponse(200, "success."), nil
}

type ClientEncodingFormat interface {
	EncodingFormat() string
}

func (g *GB28181) SendMessage(client server.Client, data any) (message.Response, error) {
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
		message.NewAllowHeader(),
	)

	if en, ok := client.(ClientEncodingFormat); ok {
		switch strings.ToLower(en.EncodingFormat()) {
		case "gb2312", "gbk":
			msg.SetBody(string(message.ContentType__MANSCDP_XML), append([]byte("<?xml version=\"1.0\" encoding=\"GBK\"?>\r\n"), content...))
		default:
			msg.SetBody(string(message.ContentType__MANSCDP_XML), append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n"), content...))
		}
	} else {
		msg.SetBody(string(message.ContentType__MANSCDP_XML), content)
	}

	return g.server.SendMessage(client, msg.(message.Request))
}

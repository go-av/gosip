package gb28181

import (
	"context"
	"fmt"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type MediaServer interface {
}

func (g *GB28181) Invite(ctx context.Context, client server.Client, deviceID string, sdp string) (dialog.Dialog, error) {
	protocol, address := client.Transport()
	server := g.server.ServerAddress()
	fromAddr := &utils.HostAndPort{
		Host: server.Host,
		Port: server.Port,
	}
	from := dialog.NewFrom("", g.handler.ServerSIPID(), protocol, fromAddr.String())
	to := dialog.NewTo(deviceID, address)
	return g.server.Invite(ctx, from, to, sdp, func(msg message.Message) {
		msg.SetHeader(message.NewSubjectHeader(fmt.Sprintf("%s:%d,%s:%d", deviceID, 1200010001, client.User(), 0)))
	})
}

func (g *GB28181) Bye(client server.Client, deviceID string, callID string) error {
	protocol, address := client.Transport()
	hostAndPort, _ := utils.ParseHostAndPort(address)
	clientAddress := message.NewAddress(client.User(), hostAndPort.Host, hostAndPort.Port).WithDomain(g.handler.Realm())
	msg := message.NewRequestMessage(protocol, method.BYE, clientAddress)
	msg.AppendHeader(
		message.NewViaHeader(protocol, g.server.ServerAddress().Host, g.server.ServerAddress().Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewCSeqHeader(1, method.BYE),
		message.NewFromHeader("", g.server.ServerAddress().Clone().SetUser(g.handler.ServerSIPID()).WithDomain(g.handler.Realm()), message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader("", clientAddress, nil),
		message.NewMaxForwardsHeader(70),
		message.NewAllowHeader(),
		message.NewCallIDHeader(callID),
	)

	return g.server.Send(protocol, address, msg.(message.Request))
}

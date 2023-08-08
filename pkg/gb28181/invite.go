package gb28181

import (
	"context"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

func (g *GB28181) Invite(ctx context.Context, client server.Client, deviceID string, sdp string, header func(msg message.Message)) (dialog.Dialog, error) {
	protocol, address := client.Transport()
	server := g.server.ServerAddress()
	fromAddr := &utils.HostAndPort{
		Host: server.Host,
		Port: server.Port,
	}
	from := dialog.NewFrom("", g.handler.ServerSIPID(), protocol, fromAddr.String())
	to := dialog.NewTo(deviceID, address)
	return g.server.Invite(ctx, from, to, sdp, header)
}

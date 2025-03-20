package sip

import (
	"context"

	"github.com/go-av/gosip/pkg/gb28181"
	"github.com/go-av/gosip/pkg/server"
)

func (d *SipHandler) Keepalive(ctx context.Context, client server.Client, msg *gb28181.Keepalive) (*server.Response, error) {
	return server.NewResponse(200, "OK"), nil
}

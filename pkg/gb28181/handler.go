package gb28181

import (
	"context"

	"github.com/go-av/gosip/pkg/server"
)

type GB28181Handler interface {
	Keepalive(context.Context, server.Client, *Keepalive) (*server.Response, error)
	DeviceInfo(ctx context.Context, client server.Client, msg *DeviceInfo) (*server.Response, error)
	DeviceStatus(ctx context.Context, client server.Client, msg *DeviceStatus) (*server.Response, error)
	PresetQuery(ctx context.Context, client server.Client, msg *PresetQuery) (*server.Response, error)
	Catalog(context.Context, server.Client, *Catalog) error
	Realm() string
	ServerSIPID() string
	Broadcast(ctx context.Context, client server.Client, msg *BroadcastResponse)
}

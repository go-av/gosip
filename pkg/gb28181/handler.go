package gb28181

import "github.com/go-av/gosip/pkg/server"

type GB28181Handler interface {
	Keepalive(server.Client, *Keepalive) (*server.Response, error)
	DeviceInfo(client server.Client, msg *DeviceInfo) (*server.Response, error)
	DeviceStatus(client server.Client, msg *DeviceStatus) (*server.Response, error)
	PresetQuery(client server.Client, msg *PresetQuery) (*server.Response, error)
	Catalog(server.Client, *Catalog) error
	Realm() string
	ServerSIPID() string
	Broadcast(client server.Client, msg *BroadcastResponse)
}

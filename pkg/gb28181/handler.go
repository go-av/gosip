package gb28181

import "github.com/go-av/gosip/pkg/server"

type GB28181Handler interface {
	Keepalive(*Keepalive) (*server.Response, error)
	DeviceInfo(msg *DeviceInfo) (*server.Response, error)
	DeviceStatus(msg *DeviceStatus) (*server.Response, error)
	PresetQuery(msg *PresetQuery) (*server.Response, error)
	Catalog(*Catalog) error
	Realm() string
	ServerSIPID() string
}

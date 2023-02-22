package controller

import (
	"fmt"
	"sync"

	"github.com/go-av/gosip/pkg/gb28181"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/server"
)

func NewHandler(sipID string, realm string) *ServerHandler {
	handler := &ServerHandler{
		sipID:   sipID,
		realm:   realm,
		clients: &sync.Map{},
		gb28181: &gb28181.GB28181{},
	}
	return handler
}

type ServerHandler struct {
	sipID   string
	realm   string
	server  server.Server
	clients *sync.Map
	gb28181 *gb28181.GB28181
}

func (handler *ServerHandler) SetServer(s server.Server) {
	handler.server = s
	handler.gb28181 = gb28181.NewGB28181(s, handler)
}

func (d *ServerHandler) GetClient(user string) (server.Client, error) {
	client, ok := d.clients.LoadOrStore(user, &Client{
		user:   user,
		auth:   false,
		server: d,
	})
	if !ok {
		fmt.Println("发现新用户-----", user)
	}
	return client.(*Client), nil
}

func (d *ServerHandler) Realm() string {
	return d.realm
}

func (d *ServerHandler) ServerSIPID() string {
	return d.sipID
}

func (d *ServerHandler) ReceiveMessage(body message.Body) (*server.Response, error) {
	if body.ContentType() == "Application/MANSCDP+xml" {
		return d.gb28181.Handler(body.Data())
	}
	return server.NewResponse(400, "unknown"), nil
}

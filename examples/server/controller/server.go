package controller

import (
	"fmt"
	"sync"

	"github.com/go-av/gosip/pkg/server"
)

func NewServer(realm string) *Server {
	return &Server{
		realm:   realm,
		clients: &sync.Map{},
	}
}

type Server struct {
	realm   string
	server  server.Server
	clients *sync.Map
}

func (d *Server) SetServer(s server.Server) {
	d.server = s
}

func (d *Server) GetClient(user string) (server.Client, error) {
	client, ok := d.clients.LoadOrStore(user, &Client{
		user:   user,
		auth:   false,
		server: d.server,
	})
	if !ok {
		fmt.Println("新用户", user)
	}
	return client.(*Client), nil
}

func (d *Server) Realm() string {
	return d.realm
}

func (d *Server) ReceiveMessage(content server.Content) (int, string) {
	fmt.Println("ReceiveMessage", content.ContentType())
	fmt.Println("ReceiveMessage", string(content.Data()))
	return 200, "Ok"
}

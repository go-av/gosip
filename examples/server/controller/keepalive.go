package controller

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-av/gosip/pkg/gb28181"
	"github.com/go-av/gosip/pkg/server"
)

func (d *ServerHandler) Keepalive(msg *gb28181.Keepalive) (*server.Response, error) {
	fmt.Println("xxxxxxxxxxxxxx")
	spew.Dump(msg)
	fmt.Println("xxxxxxxxxxxxxx")

	return server.NewResponse(200, "Success"), nil
}

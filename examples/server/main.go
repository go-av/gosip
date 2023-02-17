package main

import (
	"context"
	"fmt"

	"github.com/go-av/gosip/examples/server/controller"
	"github.com/go-av/gosip/pkg/server"
)

func main() {
	ctx := context.Background()
	s := controller.NewServer("99999999")
	server := server.NewServer(s)

	err := server.ListenUDPServer(ctx, "172.20.30.57:25060", []string{"udp", "tcp"})
	if err != nil {
		fmt.Println(err)
	}
}

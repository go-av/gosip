package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

func main() {

	fmt.Println(utils.LocalIp())
	ctx := context.Background()
	server := server.NewServer()
	err := server.ListenUDPServer(ctx, ":5060", []string{"udp", "tcp"})
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(10 * time.Minute)

}

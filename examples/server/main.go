package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/go-av/gosip/examples/server/controller"
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

func main() {
	ip := flag.String("ip", utils.LocalIp(), "监听地址")
	port := flag.Uint64("port", 25060, "监听端口")
	flag.Parse()
	ctx := context.Background()
	s := controller.NewHandler("99999999")
	server := server.NewServer(s)
	err := server.ListenUDPServer(ctx, "0.0.0.0", *ip, uint16(*port), []string{"udp", "tcp"})
	if err != nil {
		fmt.Println(err)
	}
}

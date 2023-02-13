package main

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

func main() {
	str := "127.0.0.1:6059"
	a, err := netip.ParseAddrPort(str)
	fmt.Println(err)
	fmt.Println(a.Addr())
	fmt.Println(a.Port())
	return
	fmt.Println(utils.LocalIp())
	ctx := context.Background()
	server := server.NewServer()
	err = server.ListenUDPServer(ctx, ":5060", []string{"udp", "tcp"})
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(10 * time.Minute)

}

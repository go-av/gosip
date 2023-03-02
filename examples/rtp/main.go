package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/go-av/gosip/pkg/utils"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

var addrMap = map[string]bool{}

func setBuf() {

}

func main() {
	addr := ":40026"
	tcp, err := reuse.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listen TCP", addr)

	buf := make([]byte, 30000)
	go func() {
		for {
			conn, err := tcp.Accept()
			if err != nil {
				fmt.Println(err)
			}
			go func(conn net.Conn) {
				for {
					n, err := conn.Read(buf)
					if err == io.EOF {
						return
					}
					fmt.Println("TCP", time.Now().Format("2006-01-02 15:04:05"), conn.RemoteAddr(), ">>>", conn.LocalAddr(), n)
				}
			}(conn)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	udpConn, err := reuse.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listen UDP", addr)

	for {
		n, a, err := udpConn.ReadFrom(buf)
		if err == io.EOF {
			return
		}

		fmt.Println("UDP", time.Now().Format("2006-01-02 15:04:05"), a.String(), ">>>", n)
		aa, _ := utils.ParseHostAndPort(a.String())
		if aa.Host == "172.20.30.61" || aa.Host == "172.20.30.54" {
			conn, err := reuse.Dial("udp", "172.20.30.61"+addr, a.String())
			if err != nil {
				panic("172.20.30.61" + addr)
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				logrus.Error(err)
			}
			_ = conn.Close()
			fmt.Println("UDP", time.Now().Format("2006-01-02 15:04:05"), "172.20.30.61"+addr, ">>>", n)
		}

	}
}

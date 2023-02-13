package transport

import (
	"fmt"
	"net"
	"time"

	"github.com/go-av/gosip/pkg/message"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

type UDPTransport struct {
	host             string
	port             int
	address          *net.UDPAddr
	Connection       net.PacketConn
	transportChannel chan message.Message
}

func (ut *UDPTransport) Read() (message.Message, error) {
	buffer := make([]byte, 2048)
	n, addr, err := ut.Connection.ReadFrom(buffer)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	// logrus.Debugf("%s --> %s    %s", addr.String(), ut.address.String(), string(buffer[:n]))
	fmt.Println("[GOSIP][UDP]", time.Now().Format(time.RFC3339), addr.String(), "  -> ", ut.address.String(), "\n", string(buffer[:n]))
	return message.Parse(buffer[:n])
}

func (ut *UDPTransport) GetHost() string {
	return ut.host
}

func (ut *UDPTransport) GetPort() int {
	return ut.port
}

func (ut *UDPTransport) Build(addr string) error {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	ut.address = a

	ut.Connection, err = reuse.ListenPacket("udp", addr)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (ut *UDPTransport) Start() {
	logrus.Info("Starting UDP Listening Point ")
	for {
		msg, err := ut.Read()
		if err == nil {
			ut.transportChannel <- msg
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (ut *UDPTransport) SetTransportChannel(channel chan message.Message) {
	ut.transportChannel = channel
}

func (ut *UDPTransport) Send(address string, msg message.Message) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println(err)
		logrus.Error(err)
		return err
	}

	fmt.Println("[GOSIP][UDP]", time.Now().Format(time.RFC3339), ut.address.String(), " -> ", addr.String(), "\n", msg.String())

	conn, err := reuse.Dial("udp", ut.address.String(), addr.String())
	if err != nil {
		logrus.Error("Some error %v", err)
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg.String()))
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

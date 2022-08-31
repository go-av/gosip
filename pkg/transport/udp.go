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
	address          net.UDPAddr
	Connection       net.PacketConn
	transportChannel chan message.Message
}

func (ut *UDPTransport) Read() (message.Message, error) {
	buffer := make([]byte, 2048)
	n, _, err := ut.Connection.ReadFrom(buffer)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	logrus.Debug(string(buffer[:n]))
	return message.Parse(buffer[:n])
}

func (ut *UDPTransport) GetHost() string {
	return ut.host
}

func (ut *UDPTransport) GetPort() int {
	return ut.port
}

func (ut *UDPTransport) Build(host string, port int) error {
	ut.host = host
	ut.port = port
	ut.address = net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	var err error
	ut.Connection, err = reuse.ListenPacket("udp", ut.address.String())
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

func (ut *UDPTransport) Send(host string, port string, msg message.Message) error {
	addr, err := net.ResolveUDPAddr("udp", host+":"+port)
	if err != nil {
		fmt.Println(err)
		logrus.Error(err)
		return err
	}

	logrus.Debug(msg.String() + " -> " + host + ":" + port)
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

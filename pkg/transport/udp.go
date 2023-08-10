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
	buffer           []byte
}

func (ut *UDPTransport) Read() (message.Message, error) {
	n, addr, err := ut.Connection.ReadFrom(ut.buffer)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	// logrus.Debugf("%s --> %s    %s", addr.String(), ut.address.String(), string(buffer[:n]))
	fmt.Printf("\n\n\n[GOSIP][UDP] %s %s <<--- %s \n%s\n", time.Now().Format(time.RFC3339), ut.address.String(), addr.String(), string(ut.buffer[:n]))

	msg, err := message.Parse(ut.buffer[:n])
	if err != nil {
		return nil, err
	}
	if req, ok := msg.(message.Request); ok {
		req.SetRequestFrom("udp", addr.String())
	}
	return msg, nil
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
	ut.buffer = make([]byte, 20480)
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
		logrus.Error(err)
		return err
	}

	fmt.Printf("\n\n\n[GOSIP][UDP] %s %s --->> %s \n%s\n", time.Now().Format(time.RFC3339), ut.address.String(), addr.String(), msg.String())

	conn, err := reuse.Dial("udp", ut.address.String(), addr.String())
	if err != nil {
		logrus.Errorf("Some error %v", err)
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

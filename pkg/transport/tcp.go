package transport

import (
	"net"
	"os"

	"github.com/go-av/gosip/pkg/message"
	"github.com/sirupsen/logrus"
)

type TCPTransport struct {
	listener         *net.TCPListener
	transportChannel chan message.Message
	connTable        map[string]net.Conn
}

func (tt *TCPTransport) Read() (message.Message, error) {
	buffer := make([]byte, 2048)
	conn, err := tt.listener.Accept()
	tt.connTable[conn.RemoteAddr().String()] = conn
	if err != nil {
		logrus.Error(err)
	}
	n, err := conn.Read(buffer)
	if err != nil {
		logrus.Error(err)
	}

	return message.Parse(buffer[:n])
}

func (tt *TCPTransport) Start() {
	logrus.Info("Starting TCP Listening Point ")
	for {
		msg, err := tt.Read()
		if err == nil {
			tt.transportChannel <- msg
		}
	}
}

func (tt *TCPTransport) SetTransportChannel(channel chan message.Message) {
	tt.transportChannel = channel
}

func (tt *TCPTransport) Build(host string, port int) {
	var err error
	tcpAddr := net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	tt.connTable = make(map[string]net.Conn)
	tt.listener, err = net.ListenTCP("tcp", &tcpAddr)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

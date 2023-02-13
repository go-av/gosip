package transport

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-av/gosip/pkg/message"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

type TCPTransport struct {
	addr             *net.TCPAddr
	listener         net.Listener
	transportChannel chan message.Message
	connTable        *sync.Map
}

func (tt *TCPTransport) readConn(conn net.Conn) error {
	buffer := make([]byte, 2048)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return err
			}
			logrus.Error(err)
			return err
		}
		fmt.Println("[GOSIP][TCP]", time.Now().Format(time.RFC3339), conn.RemoteAddr().String(), "  -> ", conn.LocalAddr().String(), "\n", string(buffer[:n]))
		msg, err := message.Parse(buffer[:n])
		if err != nil {
			logrus.Error(err)
		}
		if msg != nil {
			tt.transportChannel <- msg
		}
	}
}
func (tt *TCPTransport) Read() error {
	conn, err := tt.listener.Accept()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if _, ok := tt.connTable.Load(conn.RemoteAddr().String()); !ok {
		tt.connTable.Store(conn.RemoteAddr().String(), conn)
	}

	go tt.readConn(conn)
	return nil
}

func (tt *TCPTransport) Start() {
	logrus.Info("Starting TCP Listening Point")
	for {
		tt.Read()
	}
}

func (tt *TCPTransport) SetTransportChannel(channel chan message.Message) {
	tt.transportChannel = channel
}

func (tt *TCPTransport) Build(addr string) error {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	tt.addr = a
	tt.connTable = &sync.Map{}
	tt.listener, err = reuse.Listen("tcp", a.String())
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return nil
}

func (tt *TCPTransport) Send(address string, msg message.Message) error {
	fmt.Println("[GOSIP][TCP]", time.Now().Format(time.RFC3339), tt.listener.Addr().String(), "-> ", address, "\n", msg.String())
	conn, ok := tt.connTable.Load(address)
	if !ok {
		baseConn, err := reuse.Dial("udp", tt.listener.Addr().String(), address)
		if err != nil {
			return err
		}
		conn = baseConn
		go tt.readConn(baseConn)
		tt.connTable.Store(address, baseConn)

	}

	_, err := conn.(net.Conn).Write([]byte(msg.String()))
	if err != nil {
		return err
	}

	return nil
}

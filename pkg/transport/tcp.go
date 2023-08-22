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
	connTable        sync.Map
	buffer           []byte
}

func (tt *TCPTransport) readConn(addr string, conn net.Conn) error {
	defer tt.connTable.Delete(addr)

	for {
		n, err := conn.Read(tt.buffer)
		if err != nil {
			if err == io.EOF {
				return err
			}
			logrus.Error(err)
			return err
		}
		fmt.Printf("\n\n\n[GOSIP][TCP] %s %s <<-- %s \n %s\n", time.Now().Format(time.RFC3339), conn.LocalAddr().String(), conn.RemoteAddr().String(), string(tt.buffer[:n]))
		msg, err := message.Parse(tt.buffer[:n])
		if err != nil {
			logrus.Error(err)
		}
		if msg != nil {
			if req, ok := msg.(message.Request); ok {
				req.SetRequestFrom("tcp", conn.RemoteAddr().String())
			}
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

	go tt.readConn(conn.RemoteAddr().String(), conn)
	return nil
}

func (tt *TCPTransport) Start() {
	tt.buffer = make([]byte, 20480)
	logrus.Info("Starting TCP Listening Point")
	for {
		tt.Read()
		time.Sleep(100 * time.Millisecond)
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
	tt.listener, err = reuse.Listen("tcp", a.String())
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return nil
}

func (tt *TCPTransport) Send(address string, msg message.Message) error {
	fmt.Printf("\n\n\n[GOSIP][TCP] %s %s -->> %s \n %s\n", time.Now().Format(time.RFC3339), tt.listener.Addr().String(), address, msg.String())
	conn, ok := tt.connTable.Load(address)
	if !ok {
		baseConn, err := reuse.Dial("tcp", tt.listener.Addr().String(), address)
		if err != nil {
			return err
		}
		conn = baseConn
		go tt.readConn(address, baseConn)
		tt.connTable.Store(address, baseConn)
	}

	_, err := conn.(net.Conn).Write([]byte(msg.String()))
	if err != nil {
		return err
	}

	return nil
}

package transport

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func NewTransportListenPoint(protocol string, addr string, funcs ...ListenOptionFunc) (ListeningPoint, error) {
	protocol = strings.ToLower(protocol)
	switch protocol {
	case "udp":
		logrus.Info("Creating UDP listening point")
		listner := new(UDPTransport)
		logrus.Info("Binding to " + addr)
		err := listner.Listen(addr, funcs...)
		if err != nil {
			return nil, err
		}
		return listner, nil
	case "tcp":
		logrus.Info("Creating TCP listening point")
		listner := new(TCPTransport)
		logrus.Info("Binding to " + addr)
		err := listner.Listen(addr, funcs...)
		if err != nil {
			return nil, err
		}
		return listner, nil
	// case "tls":
	// 	logrus.Info("Creating TLS listening point")
	// 	listner := new(TLSTransport)
	// 	logrus.Info("Binding to " + addr)
	// 	err := listner.Listen(addr, funcs...)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return listner, nil
	default:
		logrus.Info("Unknown protocol specified")
		panic("Unknown protocol specified")
	}
}

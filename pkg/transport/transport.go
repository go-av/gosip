package transport

import (
	"github.com/sirupsen/logrus"
)

func NewTransportListenPoint(protocol string, addr string) (ListeningPoint, error) {
	switch protocol {
	case "udp":
		logrus.Info("Creating UDP listening point")
		listner := new(UDPTransport)
		logrus.Info("Binding to " + addr)
		err := listner.Build(addr)
		if err != nil {
			return nil, err
		}
		return listner, nil
	case "tcp":
		logrus.Info("Creating TCP listening point")
		listner := new(TCPTransport)
		logrus.Info("Binding to " + addr)
		err := listner.Build(addr)
		if err != nil {
			return nil, err
		}
		return listner, nil
	default:
		logrus.Info("Unknown protocol specified")
		panic("Unknown protocol specified")
	}

}

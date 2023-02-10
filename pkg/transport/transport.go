package transport

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

func NewTransportListenPoint(protocol string, host string, port int) ListeningPoint {
	switch protocol {
	case "udp":
		logrus.Info("Creating UDP listening point")
		listner := new(UDPTransport)
		logrus.Info("Binding to " + host + ":" + strconv.Itoa(port))
		listner.Build(host, port)
		return listner
	case "tcp":
		logrus.Info("Creating TCP listening point")
		listner := new(TCPTransport)
		logrus.Info("Binding to " + host + ":" + strconv.Itoa(port))
		listner.Build(host, port)
		return listner
	default:
		logrus.Info("Unknown protocol specified")
		panic("Unknown protocol specified")

	}

}

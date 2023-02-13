package utils

import (
	"net"

	"github.com/sirupsen/logrus"
)

func LocalIp() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		logrus.Warnf("get address failure:err:%v", err.Error())
	}
	var ip = "localhost"
	for _, address := range addresses {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				return ip
			}
		}
	}
	return ip
}

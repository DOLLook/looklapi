package utils

import (
	"net"
	"strings"
)

var hostIp string

// 获取本机ip
func HostIp() string {
	if !IsEmpty(hostIp) {
		return hostIp
	}

	if netItfs, err := net.Interfaces(); err == nil {
		for _, itf := range netItfs {
			if (itf.Flags&net.FlagUp) != 0 && !strings.Contains(itf.Name, "vEthernet") {
				if addrs, ept := itf.Addrs(); ept == nil {
					for _, addr := range addrs {
						if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsMulticast() && !ipnet.IP.IsLinkLocalUnicast() && !ipnet.IP.IsLinkLocalMulticast() && ipnet.IP.To4() != nil {
							hostIp = ipnet.IP.String()
							return hostIp
						}
					}
				}
			}
		}
	}

	return hostIp
}

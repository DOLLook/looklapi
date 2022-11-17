package utils

import (
	"context"
	"math/big"
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

// http客户端ip
const HttpRemoteAddr = "Http-Remote-Addr"

// 获取客户端ip
func GetClientIp(ctx context.Context) string {
	x_forwarded_for_ip := GetXForwardedForIp(ctx)
	if !IsEmpty(x_forwarded_for_ip) {
		return x_forwarded_for_ip
	}

	return GetRemoteAddr(ctx)
}

// 获取X-Forwarded-For客户端ip
func GetXForwardedForIp(ctx context.Context) string {
	header := GetHttpHeader(ctx)
	if header != nil {
		ips := header.Get("X-Forwarded-For")
		if !IsEmpty(ips) {
			return strings.Split(ips, ",")[0]
		}
	}

	return ""
}

// 获取RemoteAddr客户端ip
func GetRemoteAddr(ctx context.Context) string {
	store := GetHttpCtxStore(ctx)
	if store == nil {
		return ""
	}

	if ip, ok := store.Get(HttpRemoteAddr).(string); ok {
		portIndex := strings.LastIndex(ip, ":")
		if portIndex < 0 {
			return ip
		} else {
			return ip[:portIndex]
		}
	} else {
		return ""
	}
}

// ip转long
func IpToLong(ip string) int64 {
	netIp := net.ParseIP(ip)
	if netIp == nil {
		return 0
	}

	ipL := big.NewInt(0)
	return ipL.SetBytes(netIp).Int64()
}

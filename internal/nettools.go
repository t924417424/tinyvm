package internal

import (
	"errors"
	"net"
	"strings"
)

func GenerateIP(cidr string, usedIPs []string) (string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	v4 := true
	if strings.IndexByte(cidr, '.') < 0 {
		v4 = false
	}
	// 获取CIDR的IP地址范围
	// ipRange := make([]string, 0)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		// ipRange = append(ipRange, ip.String())
		ipStr := ip.String()
		if v4 {
			if ipStr != ipnet.IP.String() && !strings.HasSuffix(ipStr, ".255") {
				if !contains(usedIPs, ip.String()) {
					return ip.String(), nil
				}
			}
		} else {
			if !strings.HasSuffix(ipStr, "::") && !strings.HasSuffix(ipStr, "::1") && !strings.HasSuffix(ipStr, "::fffe") {
				if !contains(usedIPs, ip.String()) {
					return ip.String(), nil
				}
			}
		}
	}

	return "", errors.New("无空闲IP")
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func contains(arr []string, item string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

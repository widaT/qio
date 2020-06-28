package qio

import (
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/libp2p/go-reuseport"
)

func ReuseListen(network, addr string) (net.Listener, error) {
	return reuseport.Listen(network, addr)
}

func reuseSuported() bool {
	utsname := syscall.Utsname{}
	syscall.Uname(&utsname)
	release := int8ToString(utsname.Release)
	if len(release) > 0 {
		version := strings.Split(release, "-")
		if len(version) > 1 {
			return versionSurported(version[0])
		}
	}
	return false
}

func int8ToString(x [65]int8) string {
	var buf [65]byte
	for i, b := range x {
		buf[i] = byte(b)
	}
	str := string(buf[:])
	if i := strings.Index(str, "\x00"); i != -1 {
		str = str[:i]
	}
	return str
}

//versionSurported check linux kenel verison is greater than 3.9
func versionSurported(v string) bool {
	ret := strings.Split(v, ".")
	if len(ret) < 2 {
		return false
	}
	v1, v2 := ret[0], ret[1]
	if v1 < "3" {
		return false
	}
	num, err := strconv.Atoi(v2)
	if err != nil {
		return false
	}
	if num < 9 {
		return false
	}
	return true
}

package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
)

// GetClientIP func definition
func GetClientIP(request *http.Request) (clientIP string) {
	core.Logger.Debugf("GetClientIP() - Header: %#v", request.Header)
	clientIP = request.Header.Get("Device-Ip")
	if clientIP != "" {
		return
	}
	xForwardedFor := request.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		clientIP = strings.Split(xForwardedFor, ",")[0]
		return
	}
	clientIP = request.Header.Get("X-Real-Ip")
	if clientIP != "" {
		return
	}
	clientIP = request.RemoteAddr
	return
}

// IPToInt64 func definition
func IPToInt64(ipAddress string) int64 {
	if ipAddress == "<nil>" {
		return 0
	}

	bits := strings.Split(ipAddress, ".")

	if len(bits) < 4 {
		return 0
	}

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

// Int64ToIP func definition
func Int64ToIP(ipInt int64) string {
	return fmt.Sprintf("%d.%d.%d.%d", ipInt>>24, ipInt>>16&0xff, ipInt>>8&0xff, ipInt&0xff)
}

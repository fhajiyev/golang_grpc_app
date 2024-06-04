package utils_test

import (
	"net/http"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/go-test/test"
)

var getClientIpTests = []struct {
	name     string
	headers  map[string]string
	expected string
}{
	{
		name: "Prefer Device-IP to everything else",
		headers: map[string]string{
			"Device-Ip":       "1.1.1.1",
			"X-Forwarded-For": "2.2.2.2, 3.3.3.3",
			"X-Real-Ip":       "4.4.4.4",
		},
		expected: "1.1.1.1",
	},
	{
		name: "Prefer X-Forwarded-For to X-Real-Ip",
		headers: map[string]string{
			"X-Forwarded-For": "2.2.2.2, 3.3.3.3",
			"X-Real-Ip":       "4.4.4.4",
		},
		expected: "2.2.2.2",
	},
	{
		name: "Use X-Real-IP as fallback",
		headers: map[string]string{
			"X-Real-Ip": "4.4.4.4",
		},
		expected: "4.4.4.4",
	},
}

func TestGetClientIP(t *testing.T) {
	for _, tt := range getClientIpTests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{}
			r.Header = http.Header{}
			for k, v := range tt.headers {
				r.Header.Add(k, v)
			}
			actual := utils.GetClientIP(r)
			test.AssertEqual(t, tt.expected, actual, "TestGetClientIp")
		})
	}
}

func TestIP(t *testing.T) {
	ipAddr := "220.118.92.225"
	ipInt := utils.IPToInt64(ipAddr)
	ipStr := utils.Int64ToIP(ipInt)

	t.Logf("TestIP() - ip: %v, int: %v, str: %v", ipAddr, ipInt, ipStr)

	test.AssertEqual(t, ipInt != 0, true, "TestIp")
	test.AssertEqual(t, ipAddr, ipStr, "TestIp")
}

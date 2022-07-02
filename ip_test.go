package main

import (
	"github.com/stretchr/testify/assert"
	"net"
	"net/http/httptest"
	"testing"
)

func TestParseForwardedHeader(t *testing.T) {
	header := []string{
		"for=12.34.56.78;host=example.com;proto=https, for=23.45.67.89",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=65.182.89.102",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=10.1.2.3",
		"for=192.0.2.60;proto=http;by=203.0.113.43",
		"for=10.1.2.3;proto=http;by=203.0.113.43",
		"proto=http;by=203.0.113.43;for=192.0.2.61",
		"proto=http;by=203.0.113.43;for=10.1.2.3",
		"   ",
		"",
	}
	expected := []string{
		"23.45.67.89",
		"65.182.89.102",
		"",
		"192.0.2.60",
		"",
		"192.0.2.61",
		"",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseForwardedHeader(head))
	}
}

func TestParseXForwardedForHeader(t *testing.T) {
	header := []string{
		"65.182.89.102",
		"127.0.0.1, 23.21.45.67, 65.182.89.102",
		"127.0.0.1,23.21.45.67,65.182.89.102",
		"65.182.89.102,23.21.45.67,127.0.0.1",
		"   ",
		"",
	}
	expected := []string{
		"65.182.89.102",
		"65.182.89.102",
		"65.182.89.102",
		"",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseXForwardedForHeader(head))
	}
}

func TestParseXRealIPHeader(t *testing.T) {
	header := []string{
		"",
		"  ",
		"invalid",
		"127.0.0.1",
		"65.182.89.102",
	}
	expected := []string{
		"",
		"",
		"",
		"",
		"65.182.89.102",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseXRealIPHeader(head))
	}
}

func TestGetIP(t *testing.T) {
	ipHeader = allIPHeader
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "123.456.789.012:29302"

	// no header, default
	assert.Equal(t, "123.456.789.012", getIP(r))

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	assert.Equal(t, "103.0.53.43", getIP(r))

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	assert.Equal(t, "192.0.2.60", getIP(r))

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r))

	// True-Client-IP
	r.Header.Set("True-Client-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r))

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r))

	// no parser
	ipHeader = nil
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "123.456.789.012", getIP(r))
}

func TestGetIPWithProxy(t *testing.T) {
	allowedProxySubnetList := []string{"10.0.0.0/8"}
	allowedProxySubnets := make([]net.IPNet, 0)

	for _, v := range allowedProxySubnetList {
		_, cidr, err := net.ParseCIDR(v)

		if err != nil {
			continue
		}

		allowedProxySubnets = append(allowedProxySubnets, *cidr)
	}

	allowedSubnets = allowedProxySubnets
	ipHeader = allIPHeader
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.8:29302"

	// no header, default
	assert.Equal(t, "10.0.0.8", getIP(r))

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	assert.Equal(t, "103.0.53.43", getIP(r))

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	assert.Equal(t, "192.0.2.60", getIP(r))

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r))

	// True-Client-IP
	r.Header.Set("True-Client-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r))

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r))

	// no parser
	ipHeader = nil
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "10.0.0.8", getIP(r))

	// invalid remote IP
	r.RemoteAddr = "1.1.1.1"
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "1.1.1.1", getIP(r))
}

func TestIsValidIP(t *testing.T) {
	assert.False(t, isValidIP("invalid"))
	assert.False(t, isValidIP(""))
	assert.False(t, isValidIP("  "))
	assert.False(t, isValidIP("127.0.0.1"))
	assert.False(t, isValidIP("0.0.0.0"))
	assert.True(t, isValidIP("1.2.3.4"))
}

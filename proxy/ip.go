package proxy

import (
	"net"
	"net/http"
	"strings"
)

var (
	cfConnectingIP = headerParser{"CF-Connecting-IP", parseXForwardedForHeader}
	trueClientIP   = headerParser{"True-Client-IP", parseXForwardedForHeader}
	xForwardedFor  = headerParser{"X-Forwarded-For", parseXForwardedForHeader}
	forwarded      = headerParser{"Forwarded", parseForwardedHeader}
	xRealIP        = headerParser{"X-Real-IP", parseXRealIPHeader}
	allIPHeader    = []headerParser{
		cfConnectingIP,
		trueClientIP,
		xForwardedFor,
		forwarded,
		xRealIP,
	}

	allowedSubnets []net.IPNet
	ipHeader       []headerParser
)

type parseHeaderFunc func(string) string

type headerParser struct {
	Header string
	Parser parseHeaderFunc
}

// GetIP returns the real visitor IP for the request.
func GetIP(r *http.Request) string {
	ip := cleanIP(r.RemoteAddr)

	if allowedSubnets != nil && !validProxySource(ip, allowedSubnets) {
		return ip
	}

	for _, header := range ipHeader {
		value := r.Header.Get(header.Header)

		if value != "" {
			parsedIP := header.Parser(value)

			if parsedIP != "" {
				return parsedIP
			}
		}
	}

	return ip
}

func cleanIP(ip string) string {
	if strings.Contains(ip, ":") {
		host, _, err := net.SplitHostPort(ip)

		if err != nil {
			return ip
		}

		return host
	}

	return ip
}

func validProxySource(address string, allowed []net.IPNet) bool {
	ip := net.ParseIP(address)

	if ip == nil {
		return false
	}

	for _, from := range allowed {
		if from.Contains(ip) {
			return true
		}
	}

	return false
}

func parseForwardedHeader(value string) string {
	parts := strings.Split(value, ",")

	if len(parts) > 0 {
		parts = strings.Split(parts[len(parts)-1], ";")

		for _, part := range parts {
			k, ip, found := strings.Cut(part, "=")

			if found && strings.TrimSpace(k) == "for" {
				ip = cleanIP(ip)

				if isValidIP(ip) {
					return ip
				}
			}
		}
	}

	return ""
}

func parseXForwardedForHeader(value string) string {
	parts := strings.Split(value, ",")

	if len(parts) > 0 {
		ip := cleanIP(strings.TrimSpace(parts[len(parts)-1]))

		if isValidIP(ip) {
			return ip
		}
	}

	return ""
}

func parseXRealIPHeader(value string) string {
	value = cleanIP(strings.TrimSpace(value))

	if isValidIP(value) {
		return value
	}

	return ""
}

func isValidIP(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil &&
		!ip.IsPrivate() &&
		!ip.IsLoopback() &&
		!ip.IsUnspecified()
}

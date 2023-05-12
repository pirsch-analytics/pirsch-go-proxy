package proxy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadIPHeader(t *testing.T) {
	config := new(Config)
	config.Network.Header = []string{"X-Forwarded-For", "x-Real-iP"}
	loadIPHeader(config)
	assert.Len(t, ipHeader, 2)
	assert.Equal(t, "X-Forwarded-For", ipHeader[0].Header)
	assert.Equal(t, "X-Real-IP", ipHeader[1].Header)
}

func TestLoadSubnets(t *testing.T) {
	config := new(Config)
	config.Network.Subnets = []string{"10.0.0.1/8", "123.56.98.42/16"}
	loadSubnets(config)
	assert.Len(t, allowedSubnets, 2)
	assert.Equal(t, "10.0.0.0/8", allowedSubnets[0].String())
	assert.Equal(t, "123.56.0.0/16", allowedSubnets[1].String())
}

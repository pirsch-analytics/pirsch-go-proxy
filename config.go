package main

import (
	"github.com/BurntSushi/toml"
	"github.com/emvi/logbuch"
	"net"
	"os"
	"strings"
)

type Config struct {
	Server  Server   `toml:"server"`
	Clients []Client `toml:"clients"`
	Network Network  `toml:"network"`
	BaseURL string   `toml:"base_url"`
}

type Server struct {
	Host         string `toml:"host"`
	WriteTimeout int    `toml:"write_timeout"`
	ReadTimeout  int    `toml:"read_timeout"`
	TLS          bool   `toml:"tls"`
	TLSCert      string `toml:"tls_cert"`
	TLSKey       string `toml:"tls_key"`
}

type Client struct {
	ID       string `toml:"id"`
	Secret   string `toml:"secret"`
	Hostname string `toml:"hostname"`
}

type Network struct {
	Header  []string `toml:"header"`
	Subnets []string `toml:"subnets"`
}

func loadConfig() {
	data, err := os.ReadFile("config.toml")

	if err != nil {
		logbuch.Fatal("Error loading configuration", logbuch.Fields{"err": err})
	}

	if err := toml.Unmarshal(data, &config); err != nil {
		logbuch.Fatal("Error loading configuration", logbuch.Fields{"err": err})
	}

	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 5
	}

	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 5
	}

	loadIPHeader()
	loadSubnets()
}

func loadIPHeader() {
	for _, header := range config.Network.Header {
		found := false

		for _, parser := range allIPHeader {
			if strings.ToLower(header) == strings.ToLower(parser.Header) {
				ipHeader = append(ipHeader, parser)
				found = true
				break
			}
		}

		if !found {
			logbuch.Fatal("Header invalid", logbuch.Fields{"header": header})
		}
	}
}

func loadSubnets() {
	for _, subnet := range config.Network.Subnets {
		_, n, err := net.ParseCIDR(subnet)

		if err != nil {
			logbuch.Fatal("Error parsing subnet", logbuch.Fields{"err": err, "subnet": subnet})
		}

		allowedSubnets = append(allowedSubnets, *n)
	}
}

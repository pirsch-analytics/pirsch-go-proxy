package proxy

import (
	"github.com/BurntSushi/toml"
	"github.com/emvi/logbuch"
	"net"
	"os"
	"strings"
)

type Config struct {
	Server             Server   `toml:"server"`
	Clients            []Client `toml:"clients"`
	Network            Network  `toml:"network"`
	BaseURL            string   `toml:"base_url"`
	BasePath           string   `toml:"base_path"`
	PageViewPath       string   `toml:"page_view_path"`
	EventPath          string   `toml:"event_path"`
	SessionPath        string   `toml:"session_path"`
	JSFilename         string   `toml:"js_filename"`
	EventsJSFilename   string   `toml:"events_js_filename"`
	SessionsJSFilename string   `toml:"sessions_js_filename"`
	ExtendedJSFilename string   `toml:"extended_js_filename"`
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
	ID     string `toml:"id"`
	Secret string `toml:"secret"`
}

type Network struct {
	Header  []string `toml:"header"`
	Subnets []string `toml:"subnets"`
}

// LoadConfig loads the configuration.
func LoadConfig() *Config {
	data, err := os.ReadFile("config.toml")

	if err != nil {
		logbuch.Fatal("Error loading configuration", logbuch.Fields{"err": err})
	}

	config := new(Config)

	if err := toml.Unmarshal(data, config); err != nil {
		logbuch.Fatal("Error loading configuration", logbuch.Fields{"err": err})
	}

	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 5
	}

	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 5
	}

	if config.BasePath == "" {
		config.BasePath = "/p"
	}

	if config.PageViewPath == "" {
		config.PageViewPath = "pv"
	}

	if config.EventPath == "" {
		config.EventPath = "e"
	}

	if config.SessionPath == "" {
		config.SessionPath = "s"
	}

	if config.JSFilename == "" {
		config.JSFilename = "p.js"
	}

	if config.EventsJSFilename == "" {
		config.EventsJSFilename = "e.js"
	}

	if config.SessionsJSFilename == "" {
		config.SessionsJSFilename = "s.js"
	}

	if config.ExtendedJSFilename == "" {
		config.ExtendedJSFilename = "ext.js"
	}

	loadIPHeader(config)
	loadSubnets(config)
	return config
}

func loadIPHeader(config *Config) {
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

func loadSubnets(config *Config) {
	for _, subnet := range config.Network.Subnets {
		_, n, err := net.ParseCIDR(subnet)

		if err != nil {
			logbuch.Fatal("Error parsing subnet", logbuch.Fields{"err": err, "subnet": subnet})
		}

		allowedSubnets = append(allowedSubnets, *n)
	}
}

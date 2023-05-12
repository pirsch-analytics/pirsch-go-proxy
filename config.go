package proxy

import (
	"github.com/BurntSushi/toml"
	"github.com/emvi/logbuch"
	"net"
	"os"
	"strings"
)

var (
	config *Config
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

// GetConfig returns the configuration.
func GetConfig() *Config {
	return config
}

// LoadConfig loads the toml configuration file.
func LoadConfig() {
	data, err := os.ReadFile("config.toml")

	if err != nil {
		logbuch.Fatal("Error loading configuration", logbuch.Fields{"err": err})
	}

	cfg := new(Config)

	if err := toml.Unmarshal(data, cfg); err != nil {
		logbuch.Fatal("Error loading configuration", logbuch.Fields{"err": err})
	}

	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 5
	}

	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 5
	}

	if cfg.BasePath == "" {
		cfg.BasePath = "/p"
	}

	if cfg.PageViewPath == "" {
		cfg.PageViewPath = "pv"
	}

	if cfg.EventPath == "" {
		cfg.EventPath = "e"
	}

	if cfg.SessionPath == "" {
		cfg.SessionPath = "s"
	}

	if cfg.JSFilename == "" {
		cfg.JSFilename = "p.js"
	}

	if cfg.EventsJSFilename == "" {
		cfg.EventsJSFilename = "e.js"
	}

	if cfg.SessionsJSFilename == "" {
		cfg.SessionsJSFilename = "s.js"
	}

	if cfg.ExtendedJSFilename == "" {
		cfg.ExtendedJSFilename = "ext.js"
	}

	loadIPHeader(cfg)
	loadSubnets(cfg)
	config = cfg
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

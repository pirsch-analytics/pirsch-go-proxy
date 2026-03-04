package proxy

import (
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	config *Config
)

type Config struct {
	Server       Server   `toml:"server"`
	Clients      []Client `toml:"clients"`
	Network      Network  `toml:"network"`
	BaseURL      string   `toml:"base_url"`
	BasePath     string   `toml:"base_path"`
	PageViewPath string   `toml:"page_view_path"`
	EventPath    string   `toml:"event_path"`
	SessionPath  string   `toml:"session_path"`
	JSFilename   string   `toml:"js_filename"`
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
	ID     string       `toml:"id"`
	Secret string       `toml:"secret"`
	Filter ClientFilter `toml:"filter"`
}

type ClientFilter struct {
	Hostname           []string `toml:"hostname"`
	Path               []string `toml:"path"`
	IdentificationCode []string `toml:"identification_code"`
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
	path := "config.toml"

	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	data, err := os.ReadFile(path)

	if err != nil {
		slog.Error("Error loading configuration", "err", err)
		panic(err)
	}

	cfg := new(Config)

	if err := toml.Unmarshal(data, cfg); err != nil {
		slog.Error("Error loading configuration", "err", err)
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
		cfg.JSFilename = "pa.js"
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

		if strings.ToLower(header) == "caddy" {
			ipHeader = append(ipHeader, xForwardedForCaddy)
			found = true
			break
		}

		if !found {
			slog.Error("Header invalid", "header", header)
			panic("Header invalid")
		}
	}
}

func loadSubnets(config *Config) {
	for _, subnet := range config.Network.Subnets {
		_, n, err := net.ParseCIDR(subnet)

		if err != nil {
			slog.Error("Error parsing subnet", "err", err, "subnet", subnet)
			return
		}

		allowedSubnets = append(allowedSubnets, *n)
	}
}

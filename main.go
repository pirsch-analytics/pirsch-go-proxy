package main

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/NYTimes/gziphandler"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	config Config
)

type Config struct {
	Server  Server   `toml:"server"`
	Clients []Client `toml:"clients"`
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

func configureLogging() {
	logbuch.SetFormatter(logbuch.NewFieldFormatter("2006-01-02_15:04:05", "\t"))
	logbuch.SetLevel(logbuch.LevelInfo)
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
}

func configureRoutes(router *mux.Router) {
	router.HandleFunc("/pirsch/hit", hit)
	router.HandleFunc("/pirsch/event", event)
	router.PathPrefix("/pirsch/").Handler(http.StripPrefix("/pirsch/", gziphandler.GzipHandler(http.FileServer(http.Dir("js")))))
}

func hit(w http.ResponseWriter, r *http.Request) {
	logbuch.Info("hit")
}

func event(w http.ResponseWriter, r *http.Request) {
	logbuch.Info("event")
}

func configureCors(router *mux.Router) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400, // one day
	})
	return c.Handler(router)
}

func startServer(handler http.Handler) {
	logbuch.Info("Starting server...", logbuch.Fields{
		"write_timeout": config.Server.WriteTimeout,
		"read_timeout":  config.Server.ReadTimeout,
	})

	server := &http.Server{
		Handler:      handler,
		Addr:         config.Server.Host,
		WriteTimeout: time.Duration(config.Server.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.Server.ReadTimeout) * time.Second,
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		logbuch.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

		if err := server.Shutdown(ctx); err != nil {
			logbuch.Fatal("Error shutting down server gracefully", logbuch.Fields{"err": err})
		}

		cancel()
	}()

	if config.Server.TLS {
		logbuch.Info("TLS enabled")

		if err := server.ListenAndServeTLS(config.Server.TLSCert, config.Server.TLSKey); err != nil && err != http.ErrServerClosed {
			logbuch.Fatal(err.Error())
		}
	} else {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logbuch.Fatal(err.Error())
		}
	}
}

func main() {
	configureLogging()
	loadConfig()
	router := mux.NewRouter()
	configureRoutes(router)
	startServer(configureCors(router))
}

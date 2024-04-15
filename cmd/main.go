package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/pirsch-analytics/pirsch-go-proxy/pkg/proxy"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func configureLogging() {
	logbuch.SetFormatter(logbuch.NewFieldFormatter("2006-01-02_15:04:05", "\t"))
	logbuch.SetLevel(logbuch.LevelInfo)
}

func logSnippets() {
	cfg := proxy.GetConfig()
	fmt.Println("\npa.js:")
	fmt.Println(fmt.Sprintf(`<script defer type="text/javascript"
	src="%s"
	id="pianjs"
	data-hit-endpoint="%s"
	data-event-endpoint="%s"
	data-session-endpoint="%s"></script>`, filepath.Join(cfg.BasePath, cfg.JSFilename), filepath.Join(cfg.BasePath, cfg.PageViewPath), filepath.Join(cfg.BasePath, cfg.EventPath), filepath.Join(cfg.BasePath, cfg.SessionPath)))
	fmt.Println()
}

func startServer(handler http.Handler) {
	cfg := proxy.GetConfig()
	logbuch.Info("Starting server...", logbuch.Fields{
		"write_timeout": cfg.Server.WriteTimeout,
		"read_timeout":  cfg.Server.ReadTimeout,
		"host":          cfg.Server.Host,
	})

	server := &http.Server{
		Handler:      handler,
		Addr:         cfg.Server.Host,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
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

	if cfg.Server.TLS {
		logbuch.Info("TLS enabled")

		if err := server.ListenAndServeTLS(cfg.Server.TLSCert, cfg.Server.TLSKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logbuch.Fatal(err.Error())
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logbuch.Fatal(err.Error())
		}
	}
}

func main() {
	configureLogging()
	proxy.LoadConfig()
	proxy.SetupClients()
	logSnippets()
	startServer(proxy.GetRouter())
}

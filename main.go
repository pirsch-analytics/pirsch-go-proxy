package main

import (
	"context"
	"encoding/json"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"github.com/klauspost/compress/gzhttp"
	"github.com/pirsch-analytics/pirsch-go-proxy/proxy"
	"github.com/pirsch-analytics/pirsch-go-sdk"
	"github.com/rs/cors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	config                                                             *proxy.Config
	clients                                                            []*pirsch.Client
	pirschJS, eventsJS, sessionsJS, extendedJS                         []byte
	updatePirschJS, updateEventsJS, updateSessionsJS, updateExtendedJS time.Time
	m                                                                  sync.RWMutex
)

func configureLogging() {
	logbuch.SetFormatter(logbuch.NewFieldFormatter("2006-01-02_15:04:05", "\t"))
	logbuch.SetLevel(logbuch.LevelInfo)
}

func setupClients() {
	for _, c := range config.Clients {
		logbuch.Info("Adding client", logbuch.Fields{"id": c.ID, "base_url": config.BaseURL})
		client := pirsch.NewClient(c.ID, c.Secret, &pirsch.ClientConfig{
			BaseURL: config.BaseURL,
		})

		if _, err := client.Domain(); err != nil {
			logbuch.Fatal("Error connecting client", logbuch.Fields{"err": err})
		}

		clients = append(clients, client)
	}
}

func configureRoutes(router *mux.Router) {
	router.HandleFunc(filepath.Join(config.BasePath, config.PageViewPath), pageView)
	router.HandleFunc(filepath.Join(config.BasePath, config.EventPath), event)
	router.HandleFunc(filepath.Join(config.BasePath, config.SessionPath), session)
	serveFile(router, config.JSFilename, "pirsch.js", &pirschJS, &updatePirschJS)
	serveFile(router, config.EventsJSFilename, "pirsch-events.js", &eventsJS, &updateEventsJS)
	serveFile(router, config.SessionsJSFilename, "pirsch-sessions.js", &sessionsJS, &updateSessionsJS)
	serveFile(router, config.ExtendedJSFilename, "pirsch-extended.js", &extendedJS, &updateExtendedJS)
}

func serveFile(router *mux.Router, filename, file string, content *[]byte, updateAt *time.Time) {
	router.HandleFunc(filepath.Join(config.BasePath, filename), gzhttp.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.RLock()

		if content == nil || updateAt.Before(time.Now()) {
			m.RUnlock()

			if err := downloadFile(file, content, updateAt); err != nil {
				logbuch.Error("Error downloading script", logbuch.Fields{"err": err, "file": file})
				w.WriteHeader(http.StatusNotFound)
				return
			}

			m.RLock()
		}

		defer m.RUnlock()

		if _, err := w.Write(*content); err != nil {
			logbuch.Error("Error sending script", logbuch.Fields{"err": err, "file": file})
		}
	})))
}

func downloadFile(file string, content *[]byte, updateAt *time.Time) error {
	m.Lock()
	defer m.Unlock()
	*updateAt = time.Now().Add(time.Hour)
	resp, err := http.Get("https://api.pirsch.io/" + file)

	if err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	*content = data
	return nil
}

func pageView(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	width, _ := strconv.ParseInt(query.Get("w"), 10, 16)
	height, _ := strconv.ParseInt(query.Get("h"), 10, 16)
	options := &pirsch.HitOptions{
		URL:            query.Get("url"),
		IP:             proxy.GetIP(r),
		UserAgent:      r.Header.Get("User-Agent"),
		AcceptLanguage: r.Header.Get("Accept-Language"),
		Title:          query.Get("t"),
		Referrer:       query.Get("ref"),
		ScreenWidth:    int(width),
		ScreenHeight:   int(height),
	}

	for _, client := range clients {
		if err := client.HitWithOptions(r, options); err != nil {
			logbuch.Error("Error sending page view", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
	}
}

func event(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e := struct {
		URL           string            `json:"url"`
		Title         string            `json:"title"`
		Referrer      string            `json:"referrer"`
		ScreenWidth   int               `json:"screen_width"`
		ScreenHeight  int               `json:"screen_height"`
		EventName     string            `json:"event_name"`
		EventDuration int               `json:"event_duration"`
		EventMeta     map[string]string `json:"event_meta"`
	}{}

	if err := json.Unmarshal(body, &e); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	options := &pirsch.HitOptions{
		URL:            e.URL,
		IP:             proxy.GetIP(r),
		UserAgent:      r.Header.Get("User-Agent"),
		AcceptLanguage: r.Header.Get("Accept-Language"),
		Title:          e.Title,
		Referrer:       e.Referrer,
		ScreenWidth:    e.ScreenWidth,
		ScreenHeight:   e.ScreenHeight,
	}

	for _, client := range clients {
		if err := client.EventWithOptions(e.EventName, e.EventDuration, e.EventMeta, r, options); err != nil {
			logbuch.Error("Error sending event", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
	}
}

func session(w http.ResponseWriter, r *http.Request) {
	for _, client := range clients {
		if err := client.Session(r); err != nil {
			logbuch.Error("Error extending session", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
	}
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
	config = proxy.LoadConfig()
	setupClients()
	router := mux.NewRouter()
	configureRoutes(router)
	startServer(configureCors(router))
}

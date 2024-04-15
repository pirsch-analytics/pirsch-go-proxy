package proxy

import (
	"encoding/json"
	"github.com/emvi/logbuch"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/klauspost/compress/gzhttp"
	pirsch "github.com/pirsch-analytics/pirsch-go-sdk/v2/pkg"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	pirschJS       []byte
	updatePirschJS time.Time
	m              sync.RWMutex
)

// GetRouter sets up and returns the router.
func GetRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400, // one day
	}))
	router.Get(filepath.Join(config.BasePath, config.PageViewPath), pageView)
	router.Post(filepath.Join(config.BasePath, config.EventPath), event)
	router.Post(filepath.Join(config.BasePath, config.SessionPath), session)
	serveScript(router, config.JSFilename, "pa.js", &pirschJS, &updatePirschJS)
	return router
}

func serveScript(router *chi.Mux, filename, file string, content *[]byte, updateAt *time.Time) {
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
	options := &pirsch.PageViewOptions{
		URL:                    query.Get("url"),
		IP:                     getIP(r),
		UserAgent:              r.Header.Get("User-Agent"),
		AcceptLanguage:         r.Header.Get("Accept-Language"),
		SecCHUA:                r.Header.Get("Sec-CH-UA"),
		SecCHUAMobile:          r.Header.Get("Sec-CH-UA-Mobile"),
		SecCHUAPlatform:        r.Header.Get("Sec-CH-UA-Platform"),
		SecCHUAPlatformVersion: r.Header.Get("Sec-CH-UA-Platform-Version"),
		SecCHWidth:             r.Header.Get("Sec-CH-Width"),
		SecCHViewportWidth:     r.Header.Get("Sec-CH-Viewport-Width"),
		Title:                  query.Get("t"),
		Referrer:               query.Get("ref"),
		ScreenWidth:            int(width),
		ScreenHeight:           int(height),
	}

	for _, client := range clients {
		if err := client.PageView(r, options); err != nil {
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

	options := &pirsch.PageViewOptions{
		URL:                    e.URL,
		IP:                     getIP(r),
		UserAgent:              r.Header.Get("User-Agent"),
		AcceptLanguage:         r.Header.Get("Accept-Language"),
		SecCHUA:                r.Header.Get("Sec-CH-UA"),
		SecCHUAMobile:          r.Header.Get("Sec-CH-UA-Mobile"),
		SecCHUAPlatform:        r.Header.Get("Sec-CH-UA-Platform"),
		SecCHUAPlatformVersion: r.Header.Get("Sec-CH-UA-Platform-Version"),
		SecCHWidth:             r.Header.Get("Sec-CH-Width"),
		SecCHViewportWidth:     r.Header.Get("Sec-CH-Viewport-Width"),
		Title:                  e.Title,
		Referrer:               e.Referrer,
		ScreenWidth:            e.ScreenWidth,
		ScreenHeight:           e.ScreenHeight,
	}

	for _, client := range clients {
		if err := client.Event(e.EventName, e.EventDuration, e.EventMeta, r, options); err != nil {
			logbuch.Error("Error sending event", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
	}
}

func session(w http.ResponseWriter, r *http.Request) {
	for _, client := range clients {
		options := &pirsch.PageViewOptions{
			IP:             getIP(r),
			UserAgent:      r.Header.Get("User-Agent"),
			AcceptLanguage: r.Header.Get("Accept-Language"),
		}

		if err := client.Session(r, options); err != nil {
			logbuch.Error("Error extending session", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
	}
}

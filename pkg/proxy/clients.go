package proxy

import (
	"log/slog"

	pirsch "github.com/pirsch-analytics/pirsch-go-sdk/v2/pkg"
)

var (
	clients []client
)

type client struct {
	api    *pirsch.Client
	filter []FilterFunc
}

// SetupClients initializes all configured clients.
func SetupClients() {
	for _, c := range config.Clients {
		slog.Info("Adding client", "id", c.ID, "base_url", config.BaseURL)
		pirschClient := pirsch.NewClient(c.ID, c.Secret, &pirsch.ClientConfig{
			BaseURL: config.BaseURL,
		})

		if c.ID != "" {
			if _, err := pirschClient.Domain(); err != nil {
				slog.Error("Error connecting client", "err", err)
				panic(err)
			}
		}

		clients = append(clients, client{
			api:    pirschClient,
			filter: createFilter(c.Filter),
		})
	}
}

func createFilter(config ClientFilter) []FilterFunc {
	f := make([]FilterFunc, 0)

	if len(config.Hostname) > 0 {
		f = append(f, NewHostnameFilter(config.Hostname))
	}

	if len(config.Path) > 0 {
		f = append(f, NewPathFilter(config.Path))
	}

	if len(config.IdentificationCode) > 0 {
		f = append(f, NewIdentificationCodeFilter(config.IdentificationCode))
	}

	return f
}

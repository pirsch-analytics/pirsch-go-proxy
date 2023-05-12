package proxy

import (
	"github.com/emvi/logbuch"
	"github.com/pirsch-analytics/pirsch-go-sdk"
)

var (
	clients []*pirsch.Client
)

// SetupClients initializes all configured clients.
func SetupClients() {
	for _, c := range config.Clients {
		logbuch.Info("Adding client", logbuch.Fields{"id": c.ID, "base_url": config.BaseURL})
		client := pirsch.NewClient(c.ID, c.Secret, &pirsch.ClientConfig{
			BaseURL: config.BaseURL,
		})

		if c.ID != "" {
			if _, err := client.Domain(); err != nil {
				logbuch.Fatal("Error connecting client", logbuch.Fields{"err": err})
			}
		}

		clients = append(clients, client)
	}
}

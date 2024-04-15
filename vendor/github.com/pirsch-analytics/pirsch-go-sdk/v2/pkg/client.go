package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	defaultBaseURL        = "https://api.pirsch.io"
	defaultTimeout        = time.Second * 5
	defaultRequestRetries = 5

	authenticationEndpoint  = "/api/v1/token"
	hitEndpoint             = "/api/v1/hit"
	eventEndpoint           = "/api/v1/event"
	sessionEndpoint         = "/api/v1/session"
	domainEndpoint          = "/api/v1/domain"
	sessionDurationEndpoint = "/api/v1/statistics/duration/session"
	timeOnPageEndpoint      = "/api/v1/statistics/duration/page"
	utmSourceEndpoint       = "/api/v1/statistics/utm/source"
	utmMediumEndpoint       = "/api/v1/statistics/utm/medium"
	utmCampaignEndpoint     = "/api/v1/statistics/utm/campaign"
	utmContentEndpoint      = "/api/v1/statistics/utm/content"
	utmTermEndpoint         = "/api/v1/statistics/utm/term"
	totalVisitorsEndpoint   = "/api/v1/statistics/total"
	visitorsEndpoint        = "/api/v1/statistics/visitor"
	pagesEndpoint           = "/api/v1/statistics/page"
	entryPagesEndpoint      = "/api/v1/statistics/page/entry"
	exitPagesEndpoint       = "/api/v1/statistics/page/exit"
	conversionGoalsEndpoint = "/api/v1/statistics/goals"
	eventsEndpoint          = "/api/v1/statistics/events"
	eventMetadataEndpoint   = "/api/v1/statistics/event/meta"
	listEventsEndpoint      = "/api/v1/statistics/event/list"
	growthRateEndpoint      = "/api/v1/statistics/growth"
	activeVisitorsEndpoint  = "/api/v1/statistics/active"
	timeOfDayEndpoint       = "/api/v1/statistics/hours"
	languageEndpoint        = "/api/v1/statistics/language"
	referrerEndpoint        = "/api/v1/statistics/referrer"
	osEndpoint              = "/api/v1/statistics/os"
	osVersionEndpoint       = "/api/v1/statistics/os/version"
	browserEndpoint         = "/api/v1/statistics/browser"
	browserVersionEndpoint  = "/api/v1/statistics/browser/version"
	countryEndpoint         = "/api/v1/statistics/country"
	cityEndpoint            = "/api/v1/statistics/city"
	platformEndpoint        = "/api/v1/statistics/platform"
	screenEndpoint          = "/api/v1/statistics/screen"
	tagKeysEndpoint         = "/api/v1/statistics/tags"
	tagDetailsEndpoint      = "/api/v1/statistics/tag/details"
	keywordsEndpoint        = "/api/v1/statistics/keywords"
)

var referrerQueryParams = []string{
	"ref",
	"referer",
	"referrer",
	"source",
	"utm_source",
}

// Client is used to access the Pirsch API.
type Client struct {
	baseURL        string
	logger         *slog.Logger
	clientID       string
	clientSecret   string
	accessToken    string
	expiresAt      time.Time
	timeout        time.Duration
	requestRetries int
	m              sync.RWMutex
}

// ClientConfig is used to configure the Client.
type ClientConfig struct {
	// BaseURL is optional and can be used to configure a different host for the API.
	// This is usually left empty in production environments.
	BaseURL string

	// Timeout is the timeout for HTTP requests. 5 seconds by default.
	Timeout time.Duration

	// RequestRetries sets the maximum number of requests before an error is returned. 5 retries by default.
	RequestRetries int

	// Logger is an optional logger for debugging.
	Logger slog.Handler
}

// PageViewOptions optional parameters to send with the hit request.
type PageViewOptions struct {
	URL                    string
	IP                     string
	UserAgent              string
	AcceptLanguage         string
	SecCHUA                string
	SecCHUAMobile          string
	SecCHUAPlatform        string
	SecCHUAPlatformVersion string
	SecCHWidth             string
	SecCHViewportWidth     string
	Title                  string
	Referrer               string
	ScreenWidth            int
	ScreenHeight           int
	Tags                   map[string]string
}

// NewClient creates a new client for given client ID, client secret, hostname, and optional configuration.
// A new client ID and secret can be generated on the Pirsch dashboard.
// The hostname must match the hostname you configured on the Pirsch dashboard (e.g. example.com).
// The clientID is optional when using single access tokens.
func NewClient(clientID, clientSecret string, config *ClientConfig) *Client {
	if config == nil {
		config = &ClientConfig{
			BaseURL: defaultBaseURL,
		}
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}

	if config.RequestRetries <= 0 {
		config.RequestRetries = defaultRequestRetries
	}

	if config.Logger == nil {
		config.Logger = slog.NewTextHandler(os.Stderr, nil)
	}

	c := &Client{
		baseURL:        config.BaseURL,
		logger:         slog.New(config.Logger),
		clientID:       clientID,
		clientSecret:   clientSecret,
		timeout:        config.Timeout,
		requestRetries: config.RequestRetries,
	}

	// single access tokens do not require to query an access token using oAuth
	if clientID == "" {
		c.accessToken = clientSecret
	}

	return c
}

// PageView sends a page hit to Pirsch for given http.Request and options.
func (client *Client) PageView(r *http.Request, options *PageViewOptions) error {
	if options == nil {
		options = new(PageViewOptions)
	}

	hit := client.getPageViewData(r, options)
	return client.performPost(client.baseURL+hitEndpoint, &hit, client.requestRetries)
}

// Event sends an event to Pirsch for given http.Request and options.
func (client *Client) Event(name string, durationSeconds int, meta map[string]string, r *http.Request, options *PageViewOptions) error {
	if r.Header.Get("DNT") == "1" {
		return nil
	}

	if options == nil {
		options = new(PageViewOptions)
	}

	return client.performPost(client.baseURL+eventEndpoint, &Event{
		Name:            name,
		DurationSeconds: durationSeconds,
		Metadata:        meta,
		PageView:        client.getPageViewData(r, options),
	}, client.requestRetries)
}

// Session keeps a session alive for the given http.Request and options.
func (client *Client) Session(r *http.Request, options *PageViewOptions) error {
	if r.Header.Get("DNT") == "1" {
		return nil
	}

	if options == nil {
		options = new(PageViewOptions)
	}

	return client.performPost(client.baseURL+sessionEndpoint, &PageView{
		URL:                    r.URL.String(),
		IP:                     client.selectField(options.IP, r.RemoteAddr),
		UserAgent:              client.selectField(options.UserAgent, r.Header.Get("User-Agent")),
		SecCHUA:                client.selectField(options.SecCHUA, r.Header.Get("Sec-CH-UA")),
		SecCHUAMobile:          client.selectField(options.SecCHUAMobile, r.Header.Get("Sec-CH-UA-Mobile")),
		SecCHUAPlatform:        client.selectField(options.SecCHUAPlatform, r.Header.Get("Sec-CH-UA-Platform")),
		SecCHUAPlatformVersion: client.selectField(options.SecCHUAPlatformVersion, r.Header.Get("Sec-CH-UA-Platform-Version")),
		SecCHWidth:             client.selectField(options.SecCHWidth, r.Header.Get("Sec-CH-Width")),
		SecCHViewportWidth:     client.selectField(options.SecCHViewportWidth, r.Header.Get("Sec-CH-Viewport-Width")),
	}, client.requestRetries)
}

// Domain returns the domain for this client.
func (client *Client) Domain() (*Domain, error) {
	domains := make([]Domain, 0, 1)

	if err := client.performGet(client.baseURL+domainEndpoint, client.requestRetries, &domains); err != nil {
		return nil, err
	}

	if len(domains) != 1 {
		return nil, errors.New("domain not found")
	}

	return &domains[0], nil
}

// SessionDuration returns the session duration grouped by day.
func (client *Client) SessionDuration(filter *Filter) ([]TimeSpentStats, error) {
	stats := make([]TimeSpentStats, 0)

	if err := client.performGet(client.getStatsRequestURL(sessionDurationEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// TimeOnPage returns the time spent on pages.
func (client *Client) TimeOnPage(filter *Filter) ([]TimeSpentStats, error) {
	stats := make([]TimeSpentStats, 0)

	if err := client.performGet(client.getStatsRequestURL(timeOnPageEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMSource returns the utm sources.
func (client *Client) UTMSource(filter *Filter) ([]UTMSourceStats, error) {
	stats := make([]UTMSourceStats, 0)

	if err := client.performGet(client.getStatsRequestURL(utmSourceEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMMedium returns the utm medium.
func (client *Client) UTMMedium(filter *Filter) ([]UTMMediumStats, error) {
	stats := make([]UTMMediumStats, 0)

	if err := client.performGet(client.getStatsRequestURL(utmMediumEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMCampaign returnst he utm campaigns.
func (client *Client) UTMCampaign(filter *Filter) ([]UTMCampaignStats, error) {
	stats := make([]UTMCampaignStats, 0)

	if err := client.performGet(client.getStatsRequestURL(utmCampaignEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMContent returns the utm content.
func (client *Client) UTMContent(filter *Filter) ([]UTMContentStats, error) {
	stats := make([]UTMContentStats, 0)

	if err := client.performGet(client.getStatsRequestURL(utmContentEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMTerm returns the utm term.
func (client *Client) UTMTerm(filter *Filter) ([]UTMTermStats, error) {
	stats := make([]UTMTermStats, 0)

	if err := client.performGet(client.getStatsRequestURL(utmTermEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// TotalVisitors returns the total visitor statistics.
func (client *Client) TotalVisitors(filter *Filter) (*TotalVisitorStats, error) {
	stats := new(TotalVisitorStats)

	if err := client.performGet(client.getStatsRequestURL(totalVisitorsEndpoint, filter), client.requestRetries, stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Visitors returns the visitor statistics grouped by day.
func (client *Client) Visitors(filter *Filter) ([]VisitorStats, error) {
	stats := make([]VisitorStats, 0)

	if err := client.performGet(client.getStatsRequestURL(visitorsEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Pages returns the page statistics grouped by page.
func (client *Client) Pages(filter *Filter) ([]PageStats, error) {
	stats := make([]PageStats, 0)

	if err := client.performGet(client.getStatsRequestURL(pagesEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// EntryPages returns the entry page statistics grouped by page.
func (client *Client) EntryPages(filter *Filter) ([]EntryStats, error) {
	stats := make([]EntryStats, 0)

	if err := client.performGet(client.getStatsRequestURL(entryPagesEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// ExitPages returns the exit page statistics grouped by page.
func (client *Client) ExitPages(filter *Filter) ([]ExitStats, error) {
	stats := make([]ExitStats, 0)

	if err := client.performGet(client.getStatsRequestURL(exitPagesEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// ConversionGoals returns all conversion goals.
func (client *Client) ConversionGoals(filter *Filter) ([]ConversionGoal, error) {
	stats := make([]ConversionGoal, 0)

	if err := client.performGet(client.getStatsRequestURL(conversionGoalsEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Events returns all events.
func (client *Client) Events(filter *Filter) ([]EventStats, error) {
	stats := make([]EventStats, 0)

	if err := client.performGet(client.getStatsRequestURL(eventsEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// EventMetadata returns the metadata values for an event and key.
func (client *Client) EventMetadata(filter *Filter) ([]EventStats, error) {
	stats := make([]EventStats, 0)

	if err := client.performGet(client.getStatsRequestURL(eventMetadataEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// ListEvents returns a list of all events including metadata.
func (client *Client) ListEvents(filter *Filter) ([]EventListStats, error) {
	stats := make([]EventListStats, 0)

	if err := client.performGet(client.getStatsRequestURL(listEventsEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Growth returns the growth rates for visitors, bounces, ...
func (client *Client) Growth(filter *Filter) (*Growth, error) {
	growth := new(Growth)

	if err := client.performGet(client.getStatsRequestURL(growthRateEndpoint, filter), client.requestRetries, growth); err != nil {
		return nil, err
	}

	return growth, nil
}

// ActiveVisitors returns the active visitors and what pages they're on.
func (client *Client) ActiveVisitors(filter *Filter) (*ActiveVisitorsData, error) {
	active := new(ActiveVisitorsData)

	if err := client.performGet(client.getStatsRequestURL(activeVisitorsEndpoint, filter), client.requestRetries, active); err != nil {
		return nil, err
	}

	return active, nil
}

// TimeOfDay returns the number of unique visitors grouped by time of day.
func (client *Client) TimeOfDay(filter *Filter) ([]VisitorHourStats, error) {
	stats := make([]VisitorHourStats, 0)

	if err := client.performGet(client.getStatsRequestURL(timeOfDayEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Languages returns language statistics.
func (client *Client) Languages(filter *Filter) ([]LanguageStats, error) {
	stats := make([]LanguageStats, 0)

	if err := client.performGet(client.getStatsRequestURL(languageEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Referrer returns referrer statistics.
func (client *Client) Referrer(filter *Filter) ([]ReferrerStats, error) {
	stats := make([]ReferrerStats, 0)

	if err := client.performGet(client.getStatsRequestURL(referrerEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// OS returns operating system statistics.
func (client *Client) OS(filter *Filter) ([]OSStats, error) {
	stats := make([]OSStats, 0)

	if err := client.performGet(client.getStatsRequestURL(osEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// OSVersions returns operating system version statistics.
func (client *Client) OSVersions(filter *Filter) ([]OSVersionStats, error) {
	stats := make([]OSVersionStats, 0)

	if err := client.performGet(client.getStatsRequestURL(osVersionEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Browser returns browser statistics.
func (client *Client) Browser(filter *Filter) ([]BrowserStats, error) {
	stats := make([]BrowserStats, 0)

	if err := client.performGet(client.getStatsRequestURL(browserEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// BrowserVersions returns browser version statistics.
func (client *Client) BrowserVersions(filter *Filter) ([]BrowserVersionStats, error) {
	stats := make([]BrowserVersionStats, 0)

	if err := client.performGet(client.getStatsRequestURL(browserVersionEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Country returns country statistics.
func (client *Client) Country(filter *Filter) ([]CountryStats, error) {
	stats := make([]CountryStats, 0)

	if err := client.performGet(client.getStatsRequestURL(countryEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// City returns city statistics.
func (client *Client) City(filter *Filter) ([]CityStats, error) {
	stats := make([]CityStats, 0)

	if err := client.performGet(client.getStatsRequestURL(cityEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Platform returns the platforms used by visitors.
func (client *Client) Platform(filter *Filter) (*PlatformStats, error) {
	platforms := new(PlatformStats)

	if err := client.performGet(client.getStatsRequestURL(platformEndpoint, filter), client.requestRetries, platforms); err != nil {
		return nil, err
	}

	return platforms, nil
}

// Screen returns the screen classes used by visitors.
func (client *Client) Screen(filter *Filter) ([]ScreenClassStats, error) {
	stats := make([]ScreenClassStats, 0)

	if err := client.performGet(client.getStatsRequestURL(screenEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// TagKeys returns a list of tag keys.
func (client *Client) TagKeys(filter *Filter) ([]TagStats, error) {
	stats := make([]TagStats, 0)

	if err := client.performGet(client.getStatsRequestURL(tagKeysEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Tags returns a list of tag values for a given tag key.
func (client *Client) Tags(filter *Filter) ([]TagStats, error) {
	stats := make([]TagStats, 0)

	if err := client.performGet(client.getStatsRequestURL(tagDetailsEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// Keywords returns the Google keywords, rank, and CTR.
func (client *Client) Keywords(filter *Filter) ([]Keyword, error) {
	stats := make([]Keyword, 0)

	if err := client.performGet(client.getStatsRequestURL(keywordsEndpoint, filter), client.requestRetries, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

func (client *Client) getPageViewData(r *http.Request, options *PageViewOptions) PageView {
	return PageView{
		URL:                    client.selectField(options.URL, r.URL.String()),
		IP:                     client.selectField(options.IP, r.RemoteAddr),
		UserAgent:              client.selectField(options.UserAgent, r.Header.Get("User-Agent")),
		AcceptLanguage:         client.selectField(options.AcceptLanguage, r.Header.Get("Accept-Language")),
		SecCHUA:                client.selectField(options.SecCHUA, r.Header.Get("Sec-CH-UA")),
		SecCHUAMobile:          client.selectField(options.SecCHUAMobile, r.Header.Get("Sec-CH-UA-Mobile")),
		SecCHUAPlatform:        client.selectField(options.SecCHUAPlatform, r.Header.Get("Sec-CH-UA-Platform")),
		SecCHUAPlatformVersion: client.selectField(options.SecCHUAPlatformVersion, r.Header.Get("Sec-CH-UA-Platform-Version")),
		SecCHWidth:             client.selectField(options.SecCHWidth, r.Header.Get("Sec-CH-Width")),
		SecCHViewportWidth:     client.selectField(options.SecCHViewportWidth, r.Header.Get("Sec-CH-Viewport-Width")),
		Title:                  options.Title,
		Referrer:               client.selectField(options.Referrer, client.getReferrerFromHeaderOrQuery(r)),
		ScreenWidth:            options.ScreenWidth,
		ScreenHeight:           options.ScreenHeight,
		Tags:                   options.Tags,
	}
}

func (client *Client) getReferrerFromHeaderOrQuery(r *http.Request) string {
	referrer := r.Header.Get("Referer")

	if referrer == "" {
		for _, param := range referrerQueryParams {
			referrer = r.URL.Query().Get(param)

			if referrer != "" {
				return referrer
			}
		}
	}

	return referrer
}

func (client *Client) refreshToken() error {
	client.m.Lock()
	defer client.m.Unlock()
	client.accessToken = ""
	client.expiresAt = time.Time{}
	body := struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{
		client.clientID,
		client.clientSecret,
	}
	bodyJson, err := json.Marshal(&body)

	if err != nil {
		return err
	}

	c := client.getHTTPClient()
	resp, err := c.Post(client.baseURL+authenticationEndpoint, "application/json", bytes.NewBuffer(bodyJson))

	if err != nil {
		return err
	}

	respJson := struct {
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{}

	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&respJson); err != nil {
		return err
	}

	client.accessToken = respJson.AccessToken
	client.expiresAt = respJson.ExpiresAt
	return nil
}

func (client *Client) performPost(url string, body interface{}, retry int) error {
	client.m.RLock()
	accessToken := client.accessToken
	client.m.RUnlock()

	if client.clientID != "" && retry > 0 && accessToken == "" {
		client.waitBeforeNextRequest(retry)

		if err := client.refreshToken(); err != nil {
			if client.logger != nil {
				client.logger.Error("error refreshing token", "err", err)
			}

			return errors.New(fmt.Sprintf("error refreshing token (attempt %d/%d): %s", client.requestRetries-retry, client.requestRetries, err))
		}

		return client.performPost(url, body, retry-1)
	}

	reqBody, err := json.Marshal(body)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))

	if err != nil {
		return err
	}

	client.m.RLock()
	req.Header.Set("Authorization", "Bearer "+client.accessToken)
	client.m.RUnlock()
	c := client.getHTTPClient()
	resp, err := c.Do(req)

	if err != nil {
		return err
	}

	// refresh access token and retry
	if client.clientID != "" && retry > 0 && resp.StatusCode != http.StatusOK {
		client.waitBeforeNextRequest(retry)

		if err := client.refreshToken(); err != nil {
			if client.logger != nil {
				client.logger.Error("error refreshing token", "err", err)
			}

			return errors.New(fmt.Sprintf("error refreshing token (attempt %d/%d): %s", client.requestRetries-retry, client.requestRetries, err))
		}

		return client.performPost(url, body, retry-1)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return client.requestError(url, resp.StatusCode, string(body))
	}

	return nil
}

func (client *Client) performGet(url string, retry int, result interface{}) error {
	client.m.RLock()
	accessToken := client.accessToken
	client.m.RUnlock()

	if client.clientID != "" && retry > 0 && accessToken == "" {
		client.waitBeforeNextRequest(retry)

		if err := client.refreshToken(); err != nil {
			if client.logger != nil {
				client.logger.Error("error refreshing token", "err", err)
			}

			return errors.New(fmt.Sprintf("error refreshing token (attempt %d/%d): %s", client.requestRetries-retry, client.requestRetries, err))
		}

		return client.performGet(url, retry-1, result)
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err
	}

	client.m.RLock()
	req.Header.Set("Authorization", "Bearer "+client.accessToken)
	client.m.RUnlock()
	req.Header.Set("Content-Type", "application/json")
	c := client.getHTTPClient()
	resp, err := c.Do(req)

	if err != nil {
		return err
	}

	// refresh access token and retry
	if client.clientID != "" && retry > 0 && resp.StatusCode != http.StatusOK {
		client.waitBeforeNextRequest(retry)

		if err := client.refreshToken(); err != nil {
			if client.logger != nil {
				client.logger.Error("error refreshing token", "err", err)
			}

			return errors.New(fmt.Sprintf("error refreshing token (attempt %d/%d): %s", client.requestRetries-retry, client.requestRetries, err))
		}

		return client.performGet(url, retry-1, result)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return client.requestError(url, resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(result); err != nil {
		return err
	}

	return nil
}

func (client *Client) getHTTPClient() http.Client {
	return http.Client{
		Timeout: client.timeout,
	}
}

func (client *Client) requestError(url string, statusCode int, body string) error {
	if body != "" {
		return errors.New(fmt.Sprintf("%s: received status code %d on request: %s", url, statusCode, body))
	}

	return errors.New(fmt.Sprintf("%s: received status code %d on request", url, statusCode))
}

func (client *Client) getStatsRequestURL(endpoint string, filter *Filter) string {
	u := fmt.Sprintf("%s%s", client.baseURL, endpoint)
	v := url.Values{}
	v.Add("id", filter.DomainID)
	v.Add("from", filter.From.Format("2006-01-02"))
	v.Add("to", filter.To.Format("2006-01-02"))
	v.Add("scale", string(filter.Scale))
	v.Add("tz", filter.Timezone)
	client.setURLParams(v, "path", filter.Path)
	client.setURLParams(v, "entry_path", filter.EntryPath)
	client.setURLParams(v, "exit_path", filter.ExitPath)
	client.setURLParams(v, "pattern", filter.Pattern)
	client.setURLParams(v, "event", filter.Event)
	client.setURLParams(v, "event_meta_key", filter.EventMetaKey)
	client.setURLParams(v, "language", filter.Language)
	client.setURLParams(v, "country", filter.Country)
	client.setURLParams(v, "city", filter.City)
	client.setURLParams(v, "referrer", filter.Referrer)
	client.setURLParams(v, "referrer_name", filter.ReferrerName)
	client.setURLParams(v, "os", filter.OS)
	client.setURLParams(v, "browser", filter.Browser)
	v.Add("platform", filter.Platform)
	client.setURLParams(v, "screen_class", filter.ScreenClass)
	client.setURLParams(v, "utm_source", filter.UTMSource)
	client.setURLParams(v, "utm_medium", filter.UTMMedium)
	client.setURLParams(v, "utm_campaign", filter.UTMCampaign)
	client.setURLParams(v, "utm_content", filter.UTMContent)
	client.setURLParams(v, "utm_term", filter.UTMTerm)
	v.Add("custom_metric_key", filter.CustomMetricKey)
	v.Add("custom_metric_type", string(filter.CustomMetricType))
	v.Add("offset", strconv.Itoa(filter.Offset))
	v.Add("limit", strconv.Itoa(filter.Limit))
	v.Add("sort", filter.Sort)
	v.Add("direction", filter.Direction)
	v.Add("search", filter.Search)

	if filter.Start > 0 {
		v.Set("start", strconv.Itoa(filter.Start))
	}

	for key, value := range filter.EventMeta {
		v.Add(fmt.Sprintf("meta_%s", key), value)
	}

	for key, value := range filter.Tags {
		v.Add(fmt.Sprintf("tag_%s", key), value)
	}

	if filter.IncludeAvgTimeOnPage {
		v.Set("include_avg_time_on_page", "true")
	} else {
		v.Set("include_avg_time_on_page", "false")
	}

	return u + "?" + v.Encode()
}

func (client *Client) setURLParams(v url.Values, param string, params []string) {
	for _, p := range params {
		v.Add(param, p)
	}
}

func (client *Client) waitBeforeNextRequest(retry int) {
	time.Sleep(time.Second * time.Duration(client.requestRetries-retry+1))
}

func (client *Client) selectField(a, b string) string {
	if a != "" {
		return a
	}

	return b
}

package proxy

import (
	"log/slog"
	"net/url"
	"regexp"
	"strings"
)

// FilterFunc is a client filter function.
// Returns true if the filter applies to the request URL.
type FilterFunc func(*url.URL) bool

// NewHostnameFilter returns a new FilterFunc filtering on the hostname.
// This function supports regex filters via the "regex:" prefix.
func NewHostnameFilter(hostnames []string) FilterFunc {
	directMatch, regexMatch := getMatchers(hostnames)
	return func(url *url.URL) bool {
		u := getURL(url)

		if u == nil {
			return false
		}

		hostname := strings.ToLower(u.Hostname())

		for _, match := range directMatch {
			if hostname == match {
				return true
			}
		}

		for _, match := range regexMatch {
			if match.MatchString(hostname) {
				return true
			}
		}

		return false
	}
}

// NewPathFilter returns a new FilterFunc filtering on the path.
// This function supports regex filters via the "regex:" prefix.
func NewPathFilter(paths []string) FilterFunc {
	directMatch, regexMatch := getMatchers(paths)
	return func(url *url.URL) bool {
		u := getURL(url)

		if u == nil {
			return false
		}

		path := strings.ToLower(u.Path)

		for _, match := range directMatch {
			if path == match {
				return true
			}
		}

		for _, match := range regexMatch {
			if match.MatchString(path) {
				return true
			}
		}

		return false
	}
}

// NewIdentificationCodeFilter returns a new FilterFunc filtering on the identification code query parameter.
func NewIdentificationCodeFilter(identificationCodes []string) FilterFunc {
	return func(url *url.URL) bool {
		id := url.Query().Get("code")

		for _, match := range identificationCodes {
			if id == match {
				return true
			}
		}

		return false
	}
}

func getMatchers(matchers []string) ([]string, []regexp.Regexp) {
	directMatch := make([]string, 0)
	regexMatch := make([]regexp.Regexp, 0)

	for _, matcher := range matchers {
		if strings.HasPrefix(matcher, "regex:") {
			r, err := regexp.Compile(strings.TrimPrefix(matcher, "regex:"))

			if err != nil {
				slog.Error("Failed to compile regex filter", "err", err, "regex", matcher)
				panic(err)
			}

			regexMatch = append(regexMatch, *r)
		} else {
			directMatch = append(directMatch, strings.ToLower(matcher))
		}
	}

	return directMatch, regexMatch
}

func getURL(url *url.URL) *url.URL {
	u, err := url.Parse(url.Query().Get("url"))

	if err != nil {
		return nil
	}

	return u
}

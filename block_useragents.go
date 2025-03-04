// Package traefik_plugin_block_useragents provides a plugin to block User-Agent based on browsers and OS.
package traefik_plugin_block_useragents

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

// BrowserConfig defines configuration for a single browser.
type BrowserConfig struct {
	Name    string `json:"name"`              // Browser name (e.g., "Chrome")
	Regex   string `json:"regex,omitempty"`   // Required: Exact regex pattern to match the browser
	Version string `json:"version,omitempty"` // Unused: Kept for compatibility but ignored
}

// Config holds the plugin configuration.
type Config struct {
	AllowedBrowsers []BrowserConfig `json:"allowedBrowsers,omitempty"` // List of browser configs
	AllowedOSTypes  []string        `json:"allowedOSTypes,omitempty"`  // Optional: List of allowed OS regex patterns
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{
		AllowedBrowsers: []BrowserConfig{},
		AllowedOSTypes:  []string{},
	}
}

// BlockUserAgents struct.
type BlockUserAgents struct {
	name           string
	next           http.Handler
	regexpsAllow   []*regexp.Regexp // Browser regex patterns
	osRegexpsAllow []*regexp.Regexp // OS regex patterns (optional)
}

// BlockUserAgentsMessage struct for logging blocked requests.
type BlockUserAgentsMessage struct {
	UserAgent  string `json:"user-agent"`
	RemoteAddr string `json:"ip"`
	Host       string `json:"host"`
	RequestURI string `json:"uri"`
}

// ValidateConfig validates the plugin configuration.
func ValidateConfig(config *Config) error {
	if len(config.AllowedBrowsers) == 0 {
		return fmt.Errorf("at least one allowed browser must be specified")
	}
	for _, bc := range config.AllowedBrowsers {
		if bc.Regex == "" {
			return fmt.Errorf("regex must be provided for browser: %s", bc.Name)
		}
	}
	return nil
}

// New creates and returns a plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	regexpsAllow := make([]*regexp.Regexp, 0)
	osRegexpsAllow := make([]*regexp.Regexp, 0)

	// Compile regex patterns for allowed browsers
	for _, bc := range config.AllowedBrowsers {
		if bc.Regex == "" {
			continue // Skip if no regex is provided
		}
		re, err := regexp.Compile(bc.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling browser regex for %s: %w", bc.Name, err)
		}
		regexpsAllow = append(regexpsAllow, re)
	}

	// Compile regex patterns for allowed OS types (if provided)
	for _, osPattern := range config.AllowedOSTypes {
		re, err := regexp.Compile(osPattern)
		if err != nil {
			return nil, fmt.Errorf("error compiling OS regex %q: %w", osPattern, err)
		}
		osRegexpsAllow = append(osRegexpsAllow, re)
	}

	return &BlockUserAgents{
		name:           name,
		next:           next,
		regexpsAllow:   regexpsAllow,
		osRegexpsAllow: osRegexpsAllow,
	}, nil
}

// ServeHTTP handles the HTTP request.
func (b *BlockUserAgents) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req == nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userAgent := req.UserAgent()
	if userAgent == "" {
		b.logBlockedRequest(req, "No User-Agent")
		res.WriteHeader(http.StatusForbidden)
		return
	}

	// Check browser patterns
	browserMatch := false
	for _, re := range b.regexpsAllow {
		if re.MatchString(userAgent) {
			browserMatch = true
			break
		}
	}
	if !browserMatch {
		b.logBlockedRequest(req, "Unsupported Browser")
		res.WriteHeader(http.StatusForbidden)
		return
	}

	// Check OS patterns if provided
	if len(b.osRegexpsAllow) > 0 {
		osMatch := false
		for _, re := range b.osRegexpsAllow {
			if re.MatchString(userAgent) {
				osMatch = true
				break
			}
		}
		if !osMatch {
			b.logBlockedRequest(req, "Unsupported OS")
			res.WriteHeader(http.StatusForbidden)
			return
		}
	}

	b.next.ServeHTTP(res, req)
}

// logBlockedRequest logs details of a blocked request.
func (b *BlockUserAgents) logBlockedRequest(req *http.Request, reason string) {
	message := &BlockUserAgentsMessage{
		UserAgent:  req.UserAgent(),
		RemoteAddr: req.RemoteAddr,
		Host:       req.Host,
		RequestURI: req.RequestURI,
	}
	jsonMessage, err := json.Marshal(message)
	if err == nil {
		log.Printf("%s: Blocked (%s) - %s", b.name, reason, jsonMessage)
	} else {
		log.Printf("%s: Blocked (%s) - %s", b.name, reason, req.UserAgent())
	}
}

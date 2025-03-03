// Package traefik_plugin_block_useragents provides a plugin to block User-Agent based on browsers and OS.
package traefik_plugin_block_useragents

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// BrowserConfig defines configuration for a single browser.
type BrowserConfig struct {
	Name    string `json:"name"`              // Browser name (e.g., "Chrome")
	Regex   string `json:"regex,omitempty"`   // Optional: Exact regex pattern
	Version string `json:"version,omitempty"` // Optional: Version for comparison (e.g., ">121")
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

// BlockUserAgents struct holds the plugin instance data.
type BlockUserAgents struct {
	name            string
	next            http.Handler
	allowedBrowsers []BrowserConfig
	osRegexpsAllow  []*regexp.Regexp
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
	return nil
}

// New creates and returns a plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	osRegexpsAllow := make([]*regexp.Regexp, 0)
	for _, osPattern := range config.AllowedOSTypes {
		re, err := regexp.Compile(osPattern)
		if err != nil {
			return nil, fmt.Errorf("error compiling OS regex %q: %w", osPattern, err)
		}
		osRegexpsAllow = append(osRegexpsAllow, re)
	}
	return &BlockUserAgents{
		name:            name,
		next:            next,
		allowedBrowsers: config.AllowedBrowsers,
		osRegexpsAllow:  osRegexpsAllow,
	}, nil
}

// ServeHTTP handles the HTTP request and applies the blocking logic.
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

	// Check browser conditions
	browserMatch := false
	for _, bc := range b.allowedBrowsers {
		if bc.Regex != "" {
			re, err := regexp.Compile(bc.Regex)
			if err != nil {
				log.Printf("Invalid regex for %s: %v", bc.Name, err)
				continue
			}
			if re.MatchString(userAgent) {
				browserMatch = true
				break
			}
		} else if bc.Version != "" && strings.HasPrefix(bc.Version, ">") {
			threshold := strings.TrimPrefix(bc.Version, ">")
			detectedVersion, found := extractBrowserVersion(userAgent, bc.Name)
			if found && versionGreaterThan(detectedVersion, threshold) {
				browserMatch = true
				break
			}
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

	// If all conditions pass, proceed to the next handler
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

// extractBrowserVersion extracts the version number following the browser name in the User-Agent string.
func extractBrowserVersion(userAgent, browser string) (string, bool) {
	// Escape the browser name to handle special characters
	escapedBrowser := regexp.QuoteMeta(browser)
	// Look for the browser name followed by a slash and a version number (digits and dots)
	re := regexp.MustCompile(escapedBrowser + `/([\d.]+)`)
	matches := re.FindStringSubmatch(userAgent)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

// versionGreaterThan compares two version strings numerically.
func versionGreaterThan(v1, v2 string) bool {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	// Pad the shorter version with zeros
	for len(parts1) < len(parts2) {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < len(parts1) {
		parts2 = append(parts2, "0")
	}
	for i := 0; i < len(parts1); i++ {
		p1, err1 := strconv.Atoi(parts1[i])
		p2, err2 := strconv.Atoi(parts2[i])
		if err1 != nil || err2 != nil {
			return false // Invalid version parts
		}
		if p1 > p2 {
			return true
		} else if p1 < p2 {
			return false
		}
	}
	return false // Versions are equal
}

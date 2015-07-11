// Package urlx parses and normalizes URLs. It can also resolve hostname to an IP address.
package urlx

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/purell"
)

// Parse parses raw URL string into the net/url URL struct.
// It uses the url.Parse() internally, but it slightly changes
// its behavior:
// 1. It forces the default scheme and port.
// 2. It favors absolute paths over relative ones, thus "example.com"
//    is parsed into url.Host instead of into url.Path.
// 3. It splits Host:Port into separate fields by default.
// 4. It lowercases the Host (not only the Scheme).
func Parse(rawURL string) (*url.URL, error) {
	// Force default http scheme, so net/url.Parse() doesn't
	// put both host and path into the (relative) path.
	if strings.Index(rawURL, "//") == 0 {
		// Leading double slashes (any scheme). Force http.
		rawURL = "http:" + rawURL
	}
	if strings.Index(rawURL, "://") == -1 {
		// Missing scheme. Force http.
		rawURL = "http://" + rawURL
	}

	// Use net/url.Parse() now.
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, &url.Error{"parse", rawURL, err}
	}
	if u.Host == "" {
		return nil, &url.Error{"parse", rawURL, errors.New("empty host")}
	}
	u.Host = strings.ToLower(u.Host)

	return u, nil
}

// SplitHostPort splits a network address of the form "host:port" into
// host and port. Unlike net.SplitHostPort(), it doesn't remove brackets
// from [IPv6] host and it accepts net/url.URL struct instead of a string.
func SplitHostPort(u *url.URL) (host, port string, err error) {
	host = u.Host

	// Find last colon.
	if i := strings.LastIndex(host, ":"); i != -1 {
		// If we're not inside [IPv6] brackets, split host:port.
		if len(host) > i && strings.Index(host[i:], "]") == -1 {
			port = host[i+1:]
			host = host[:i]
		}
	}

	// Host is required field.
	if host == "" {
		return host, port, &url.Error{"splithostport", host, errors.New("empty host")}
	}

	// Port is optional. But if it's set, is it a number?
	if port != "" {
		if _, err := strconv.Atoi(port); err != nil {
			return host, port, &url.Error{"splithostport", host, err}
		}
	}

	return host, port, nil
}

const normalizeFlags purell.NormalizationFlags = purell.FlagRemoveDefaultPort |
	purell.FlagDecodeDWORDHost | purell.FlagDecodeOctalHost | purell.FlagDecodeHexHost |
	purell.FlagRemoveUnnecessaryHostDots | purell.FlagRemoveDotSegments | purell.FlagRemoveDuplicateSlashes |
	purell.FlagUppercaseEscapes | purell.FlagDecodeUnnecessaryEscapes | purell.FlagEncodeNecessaryEscapes |
	purell.FlagSortQuery

// Normalize returns normalized URL string.
// Behavior:
// 1. Remove unnecessary host dots.
// 2. Remove default port (http://localhost:80 becomes http://localhost).
// 3. Remove duplicate slashes.
// 4. Remove unnecessary dots from path.
// 5. Sort query parameters.
// 6. Decode host IP into decimal numbers.
// 7. Handle escape values.
func Normalize(u *url.URL) (string, error) {
	if u == nil || u.Host == "" {
		return "", &url.Error{"normalize", u.String(), errors.New("empty host")}
	}

	return purell.NormalizeURL(u, normalizeFlags), nil
}

// NormalizeString returns normalized URL string.
// It's a shortcut for Parse() and Normalize() funcs.
func NormalizeString(rawURL string) (string, error) {
	u, err := Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Don't use purell.NormalizeURLString() directly,
	// we want to force behavior of our Parse() function.
	return purell.NormalizeURL(u, normalizeFlags), nil
}

// Resolve resolves the URL host to its IP address.
func Resolve(u *url.URL) (*net.IPAddr, error) {
	host, _, err := SplitHostPort(u)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// Resolve resolves the URL host to its IP address.
// It's a shortcut for Parse() and Resolve() funcs.
func ResolveString(rawURL string) (*net.IPAddr, error) {
	u, err := Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return Resolve(u)
}

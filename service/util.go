package service

import (
	"net/url"
	"strings"

	"github.com/dtynn/grpcproxy/config"
)

var urlScheme = map[bool]string{
	true:  HTTPSPrefix,
	false: HTTPPrefix,
}

const (
	Sep         = config.Sep
	Wildcard    = config.Wildcard
	HTTPSPrefix = "https://"
	HTTPPrefix  = "http://"
)

func str2NonEmptySlice(s, sep string) []string {
	return nonEmptySlice(strings.Split(s, sep))
}

func nonEmptySlice(pieces []string) []string {
	i := 0
	for i < len(pieces) {
		part := strings.TrimSpace(pieces[i])
		if part == "" {
			pieces = append(pieces[:i], pieces[i+1:]...)
			continue
		}

		pieces[i] = part
		i += 1
	}
	return pieces
}

func parseURL(rawurl string, tls bool) (*url.URL, error) {
	if !strings.HasPrefix(rawurl, HTTPPrefix) && !strings.HasPrefix(rawurl, HTTPSPrefix) {
		rawurl = urlScheme[tls] + rawurl
	}

	return url.Parse(rawurl)
}

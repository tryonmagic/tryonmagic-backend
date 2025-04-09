package utils

import (
	"encoding/base64"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// Custom error types
var (
	ErrInvalidShop = errors.New("received invalid shop argument")
	ErrInvalidHost = errors.New("received invalid host argument")
)

// Dummy decodeHost to match base64 decoding logic
func decodeHost(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// SanitizeShop returns a function that validates and sanitizes a shop string
func SanitizeShop(shop string, throwOnInvalid bool) (string, error) {
	shopUrl := shop
	domainsRegex := []string{
		"myshopify\\.com",
		"shopify\\.com",
		"myshopify\\.io",
		"shop\\.dev",
	}

	shopUrlRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-_]*\.(` + strings.Join(domainsRegex, "|") + `)[/]*$`)
	shopAdminRegex := regexp.MustCompile(`^admin\.(` + strings.Join(domainsRegex, "|") + `)/store/([a-zA-Z0-9][a-zA-Z0-9-_]*)$`)

	if shopAdminRegex.MatchString(shopUrl) {
		converted := ShopAdminUrlToLegacyUrl(shopUrl)
		if converted != nil {
			shopUrl = *converted
		} else {
			shopUrl = ""
		}
	}

	if shopUrlRegex.MatchString(shopUrl) {
		return shopUrl, nil
	}

	if throwOnInvalid {
		return "", ErrInvalidShop
	}
	return "", nil
}

// SanitizeHost returns a function that validates and sanitizes a host string
func SanitizeHost() func(host string, throwOnInvalid bool) (string, error) {
	return func(host string, throwOnInvalid bool) (string, error) {
		base64regex := regexp.MustCompile(`^[0-9a-zA-Z+/]+={0,2}$`)

		if !base64regex.MatchString(host) {
			if throwOnInvalid {
				return "", ErrInvalidHost
			}
			return "", nil
		}

		decoded, err := decodeHost(host)
		if err != nil {
			if throwOnInvalid {
				return "", ErrInvalidHost
			}
			return "", nil
		}

		parsedUrl, err := url.Parse("https://" + decoded)
		if err != nil {
			if throwOnInvalid {
				return "", ErrInvalidHost
			}
			return "", nil
		}

		originsRegex := []string{
			"myshopify\\.com",
			"shopify\\.com",
			"myshopify\\.io",
			"spin\\.dev",
			"shop\\.dev",
		}

		hostRegex := regexp.MustCompile(`\.(` + strings.Join(originsRegex, "|") + `)$`)
		if !hostRegex.MatchString(parsedUrl.Hostname()) {
			if throwOnInvalid {
				return "", ErrInvalidHost
			}
			return "", nil
		}

		return host, nil
	}
}

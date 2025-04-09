package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// Converts admin.shopify.com/store/my-shop to my-shop.myshopify.com
func ShopAdminUrlToLegacyUrl(shopAdminUrl string) *string {
	shopUrl := removeProtocol(shopAdminUrl)
	parts := strings.Split(shopUrl, ".")

	if len(parts) == 0 || parts[0] != "admin" {
		return nil
	}

	regex := regexp.MustCompile(`admin\..+/store/([^/]+)`)
	matches := regex.FindStringSubmatch(shopUrl)

	if len(matches) == 2 {
		shopName := matches[1]
		isSpinUrl := strings.Contains(shopUrl, "spin.dev/store/")
		isLocalUrl := strings.Contains(shopUrl, "shop.dev/store/")

		if isSpinUrl {
			return spinAdminUrlToLegacyUrl(shopUrl)
		} else if isLocalUrl {
			return localAdminUrlToLegacyUrl(shopUrl)
		} else {
			legacy := fmt.Sprintf("%s.myshopify.com", shopName)
			return &legacy
		}
	}

	return nil
}

// Converts my-shop.myshopify.com to admin.shopify.com/store/my-shop
func LegacyUrlToShopAdminUrl(legacyAdminUrl string) *string {
	shopUrl := removeProtocol(legacyAdminUrl)
	regex := regexp.MustCompile(`(.+)\.myshopify\.com$`)
	matches := regex.FindStringSubmatch(shopUrl)

	if len(matches) == 2 {
		shopName := matches[1]
		adminUrl := fmt.Sprintf("admin.shopify.com/store/%s", shopName)
		return &adminUrl
	} else {
		isSpinUrl := strings.HasSuffix(shopUrl, "spin.dev")
		isLocalUrl := strings.HasSuffix(shopUrl, "shop.dev")

		if isSpinUrl {
			return spinLegacyUrlToAdminUrl(shopUrl)
		} else if isLocalUrl {
			return localLegacyUrlToAdminUrl(shopUrl)
		}
	}

	return nil
}

func spinAdminUrlToLegacyUrl(shopAdminUrl string) *string {
	regex := regexp.MustCompile(`admin\.web\.(.+\.spin\.dev)/store/(.+)`)
	matches := regex.FindStringSubmatch(shopAdminUrl)

	if len(matches) == 3 {
		spinUrl := matches[1]
		shopName := matches[2]
		legacy := fmt.Sprintf("%s.shopify.%s", shopName, spinUrl)
		return &legacy
	}

	return nil
}

func localAdminUrlToLegacyUrl(shopAdminUrl string) *string {
	regex := regexp.MustCompile(`admin\.shop\.dev/store/(.+)`)
	matches := regex.FindStringSubmatch(shopAdminUrl)

	if len(matches) == 2 {
		shopName := matches[1]
		legacy := fmt.Sprintf("%s.shop.dev", shopName)
		return &legacy
	}

	return nil
}

func spinLegacyUrlToAdminUrl(legacyAdminUrl string) *string {
	regex := regexp.MustCompile(`(.+)\.shopify\.(.+\.spin\.dev)`)
	matches := regex.FindStringSubmatch(legacyAdminUrl)

	if len(matches) == 3 {
		shopName := matches[1]
		spinUrl := matches[2]
		adminUrl := fmt.Sprintf("admin.web.%s/store/%s", spinUrl, shopName)
		return &adminUrl
	}

	return nil
}

func localLegacyUrlToAdminUrl(legacyAdminUrl string) *string {
	regex := regexp.MustCompile(`(.+)\.shop\.dev$`)
	matches := regex.FindStringSubmatch(legacyAdminUrl)

	if len(matches) == 2 {
		shopName := matches[1]
		adminUrl := fmt.Sprintf("admin.shop.dev/store/%s", shopName)
		return &adminUrl
	}

	return nil
}

func removeProtocol(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")
	return url
}

package repository

import (
	"strings"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type StaticCatalog struct {
	apps    map[string]struct{}
	sources map[string]struct{}
}

var providers = map[string]string{
	"telkomsel":                 "Telkomsel",
	"xl":                        "XL",
	"indosat":                   "Indosat",
	"tri":                       "Tri",
	"smartfren":                 "Smartfren",
	"pln":                       "PLN",
	domain.ProductBPJSKesehatan: "BPJS Kesehatan",
	"bca":                       "BCA",
	"mandiri":                   "Mandiri",
	"bni":                       "BNI",
	"bri":                       "BRI",
	"indomaret":                 "Indomaret",
	"pos":                       "Pos Indonesia",
}

func NewStaticCatalog(apps, sources []string) *StaticCatalog {
	return &StaticCatalog{
		apps:    stringSet(apps),
		sources: stringSet(sources),
	}
}

func DefaultStaticCatalog(apps, sources string) *StaticCatalog {
	if strings.TrimSpace(apps) == "" {
		apps = "app_a,app_b"
	}
	return NewStaticCatalog(strings.Split(apps, ","), strings.Split(sources, ","))
}

func (c *StaticCatalog) ProductCategory(productType string) (string, bool) {
	return domain.ProductCategory(productType)
}

func (c *StaticCatalog) ValidProvider(productType, provider string) bool {
	if _, ok := c.ProductCategory(productType); !ok {
		return false
	}
	_, ok := providers[strings.ToLower(provider)]
	return ok
}

func (c *StaticCatalog) ProviderName(provider string) string {
	return providers[strings.ToLower(provider)]
}

func (c *StaticCatalog) ValidPartner(partner string) bool {
	return partner == "sample_partner"
}

func (c *StaticCatalog) AllowedSource(source string) bool {
	_, ok := c.sources[source]
	return ok
}

func (c *StaticCatalog) ValidApp(app string) bool {
	_, ok := c.apps[app]
	return ok
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

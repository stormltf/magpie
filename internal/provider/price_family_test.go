package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
)

// familyCatalog is a models.dev cache in which Kimi's model is listed as
// models.dev lists it, without a snapshot date.
func familyCatalog(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o755); err != nil {
		t.Fatal(err)
	}
	data := `{"moonshotai":{"models":{
	    "kimi-k3.1-pro":{"id":"kimi-k3.1-pro","cost":{"input":2,"output":10,"cache_read":0.2}},
	    "kimi-k3.1-pro-highspeed":{"id":"kimi-k3.1-pro-highspeed","cost":{"input":4,"output":20,"cache_read":0.2}}}}}`
	if err := os.WriteFile(catalog.CachePath(), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	catalog.Reset()
}

// A model's price is the model's, not its snapshot's: an id its vendor dates
// (Ark's kimi-k3-1-pro-261024) is priced at what models.dev lists for the
// same model, kimi-k3.1-pro, so an endpoint's calls stop counting at $0.
// Another model of the family keeps its own price.
func TestDatedIDPricedAtItsModelsList(t *testing.T) {
	priceHome(t)
	familyCatalog(t)
	if err := Save(Provider{ID: "ark", Name: "Ark", Key: "k", Chat: "https://ark.test/v3",
		Models: []string{"kimi-k3-1-pro-261024"}}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		pid, model string
		in         float64
	}{
		{"ark", "kimi-k3-1-pro-261024", 2},
		{"gone", "kimi-k3-1-pro-261024", 2},
		{"ark", "kimi-k3-1-pro-highspeed-261024", 4},
	} {
		if pr, ok := EffectivePrice(c.pid, c.model); !ok || pr.Input != c.in {
			t.Errorf("EffectivePrice(%q, %q) = %+v, %v; want $%v in", c.pid, c.model, pr, ok, c.in)
		}
	}
}

// One price the user gives a model is one price for its dates and its dotted
// spelling: what they set for kimi-k3-1-pro prices kimi-k3-1-pro-261024 too,
// while the id they priced itself still wins, and another model is untouched.
func TestUserPriceCoversTheModelsSpellings(t *testing.T) {
	priceHome(t)
	familyCatalog(t)
	if err := SetModelPrice("*/kimi-k3-1-pro", &catalog.Price{Input: 3, Output: 15, CacheRead: 0.3}); err != nil {
		t.Fatal(err)
	}
	for _, m := range []string{"kimi-k3-1-pro-261024", "kimi-k3.1-pro-260910", "Kimi-K3.1-Pro"} {
		if pr, ok := EffectivePrice("ark", m); !ok || pr.Input != 3 {
			t.Errorf("EffectivePrice(%q) = %+v, %v; want the user's $3", m, pr, ok)
		}
	}
	if pr, ok := EffectivePrice("ark", "kimi-k3-1-pro-highspeed-261024"); !ok || pr.Input != 4 {
		t.Errorf("the highspeed variant took pro's price: %+v, %v; want the maker's $4", pr, ok)
	}
	if err := SetModelPrice("*/kimi-k3-1-pro-261024", &catalog.Price{Input: 5, Output: 5, CacheRead: 0}); err != nil {
		t.Fatal(err)
	}
	if pr, ok := EffectivePrice("ark", "kimi-k3-1-pro-261024"); !ok || pr.Input != 5 {
		t.Errorf("the model's own price lost to its other spelling's: %+v, %v; want $5", pr, ok)
	}
}

// A price the user gives a provider for one spelling of its model is that
// model's price at it, and is more specific than one for every model it has.
func TestProviderPriceCoversTheModelsSpellings(t *testing.T) {
	priceHome(t)
	familyCatalog(t)
	if err := Save(Provider{ID: "ark", Name: "Ark", Key: "k", Chat: "https://ark.test/v3",
		Models: []string{"kimi-k3.1-pro"}}); err != nil {
		t.Fatal(err)
	}
	if err := SetModelPrice("ark/*", &catalog.Price{Input: 1, Output: 2, CacheRead: 0}); err != nil {
		t.Fatal(err)
	}
	if err := SetModelPrice("ark/kimi-k3.1-pro", &catalog.Price{Input: 6, Output: 30, CacheRead: 0.6}); err != nil {
		t.Fatal(err)
	}
	if pr, ok := EffectivePrice("ark", "kimi-k3-1-pro-261024"); !ok || pr.Input != 6 {
		t.Errorf("EffectivePrice = %+v, %v; want the model's $6 over the provider's $1", pr, ok)
	}
	if pr, ok := EffectivePrice("ark", "other-model-261024"); !ok || pr.Input != 1 {
		t.Errorf("EffectivePrice = %+v, %v; want the provider's $1", pr, ok)
	}
}

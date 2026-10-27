package sessions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

// A session's model is priced by its model's name, not its snapshot's: a
// dated id takes the price the user gave another spelling of the model, and
// its maker's list price when they gave none. A four-digit release is a
// model of its own and takes nothing from its family's price.
func TestPriceOfDatedIDTakesItsModelsPrice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o755)
	os.WriteFile(catalog.CachePath(), []byte(`{"moonshotai":{"id":"moonshotai","models":{
	    "kimi-k3.1-pro":{"id":"kimi-k3.1-pro","cost":{"input":2,"output":10,"cache_read":0.2}}}}}`), 0o644)
	catalog.Reset()
	t.Cleanup(catalog.Reset)

	if pr, ok := priceOf(settings.Load(), "kimi-k3-1-pro-261024"); !ok || pr.Input != 2 {
		t.Errorf("priceOf(kimi-k3-1-pro-261024) = %+v, %v; want the maker's $2", pr, ok)
	}
	if err := provider.SetModelPrice("*/kimi-k3-1-pro", &catalog.Price{Input: 3, Output: 15, CacheRead: 0.3}); err != nil {
		t.Fatal(err)
	}
	for _, m := range []string{"kimi-k3-1-pro-261024", "kimi-k3.1-pro-251017"} {
		if pr, ok := priceOf(settings.Load(), m); !ok || pr.Input != 3 {
			t.Errorf("priceOf(%q) = %+v, %v; want the user's $3", m, pr, ok)
		}
	}
	if err := provider.SetModelPrice("*/kimi-k2", &catalog.Price{Input: 9, Output: 9, CacheRead: 9}); err != nil {
		t.Fatal(err)
	}
	if pr, ok := priceOf(settings.Load(), "kimi-k2-0905"); ok {
		t.Errorf("priceOf(kimi-k2-0905) = %+v; kimi-k2's price is not its own", pr)
	}
}

package i18n

import (
	"testing"

	"github.com/mrkaak/restaurant-api/internal/models"
)

func TestResolverFallback(t *testing.T) {
	rows := []models.Translation{
		{EntityType: "product", EntityID: 1, Locale: "en", Field: "name", Value: "Baklawa"},
		{EntityType: "product", EntityID: 1, Locale: "ar", Field: "name", Value: "بقلاوة"},
		{EntityType: "product", EntityID: 1, Locale: "en", Field: "description", Value: "Sweet pastry"},
		// No Arabic description on purpose -> should fall back to English.
	}
	r := NewResolver("en", rows)

	if got := r.Field("product", 1, "name", "ar"); got != "بقلاوة" {
		t.Errorf("ar name = %q, want بقلاوة", got)
	}
	if got := r.Field("product", 1, "name", "en"); got != "Baklawa" {
		t.Errorf("en name = %q, want Baklawa", got)
	}
	// Missing Arabic description falls back to the default (English).
	if got := r.Field("product", 1, "description", "ar"); got != "Sweet pastry" {
		t.Errorf("ar desc fallback = %q, want Sweet pastry", got)
	}
	// Unknown locale with no default field returns "".
	if got := r.Field("product", 1, "name", "fr"); got != "Baklawa" {
		t.Errorf("fr name should fall back to default en = %q", got)
	}
	// Entirely unknown entity returns "".
	if got := r.Field("product", 999, "name", "en"); got != "" {
		t.Errorf("unknown entity = %q, want empty", got)
	}
}

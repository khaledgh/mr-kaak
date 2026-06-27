// Package i18n resolves per-entity field translations for a requested locale,
// falling back to the default locale so a label is never empty (plan §7).
package i18n

import "github.com/mrkaak/restaurant-api/internal/models"

// Resolver indexes a batch of translation rows for O(1) lookup while assembling
// a menu, so we translate many entities without re-querying per field.
type Resolver struct {
	defaultLocale string
	// index[entityType][entityID][locale][field] = value
	index map[string]map[uint64]map[string]map[string]string
}

// NewResolver builds an index from translation rows.
func NewResolver(defaultLocale string, rows []models.Translation) *Resolver {
	idx := make(map[string]map[uint64]map[string]map[string]string)
	for _, t := range rows {
		byID := idx[t.EntityType]
		if byID == nil {
			byID = make(map[uint64]map[string]map[string]string)
			idx[t.EntityType] = byID
		}
		byLocale := byID[t.EntityID]
		if byLocale == nil {
			byLocale = make(map[string]map[string]string)
			byID[t.EntityID] = byLocale
		}
		byField := byLocale[t.Locale]
		if byField == nil {
			byField = make(map[string]string)
			byLocale[t.Locale] = byField
		}
		byField[t.Field] = t.Value
	}
	return &Resolver{defaultLocale: defaultLocale, index: idx}
}

// Field returns the translated value for (entityType, id, field) in locale,
// falling back to the default locale, then to "" if neither exists.
func (r *Resolver) Field(entityType string, id uint64, field, locale string) string {
	if v, ok := r.lookup(entityType, id, locale, field); ok && v != "" {
		return v
	}
	if locale != r.defaultLocale {
		if v, ok := r.lookup(entityType, id, r.defaultLocale, field); ok {
			return v
		}
	}
	return ""
}

func (r *Resolver) lookup(entityType string, id uint64, locale, field string) (string, bool) {
	byID, ok := r.index[entityType]
	if !ok {
		return "", false
	}
	byLocale, ok := byID[id]
	if !ok {
		return "", false
	}
	byField, ok := byLocale[locale]
	if !ok {
		return "", false
	}
	v, ok := byField[field]
	return v, ok
}

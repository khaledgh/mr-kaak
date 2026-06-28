package services

import (
	"context"
	"log/slog"
	"strings"

	"github.com/mrkaak/restaurant-api/internal/i18n"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/search"
	"github.com/mrkaak/restaurant-api/pkg/logger"
)

// SearchService serves product search. It prefers Meilisearch (fast, typo
// tolerant) and falls back to a DB query when Meilisearch is unavailable, so
// search works before the Meilisearch container is up (plan §3.5).
type SearchService struct {
	meili         *search.Client
	catalog       *repository.CatalogRepo
	translations  *repository.TranslationRepo
	defaultLocale string
	baseURL       string
}

func NewSearchService(meili *search.Client, catalog *repository.CatalogRepo, tr *repository.TranslationRepo, defaultLocale, baseURL string) *SearchService {
	return &SearchService{meili: meili, catalog: catalog, translations: tr, defaultLocale: defaultLocale, baseURL: strings.TrimRight(baseURL, "/")}
}

// SearchResult is a lightweight, locale-resolved search hit.
type SearchResult struct {
	ID        uint64 `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	PriceFrom int64  `json:"price_from"`
	ImageURL  string `json:"image_url,omitempty"`
	Source    string `json:"-"`
}

// Search runs a query for a locale, returning at most limit results.
func (s *SearchService) Search(ctx context.Context, query, locale string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []SearchResult{}, nil
	}
	if locale == "" {
		locale = s.defaultLocale
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Prefer Meilisearch; on any error degrade to the DB.
	if s.meili != nil {
		if hits, err := s.meili.Search(ctx, query, true, limit); err == nil {
			return s.mapMeiliHits(hits, locale), nil
		} else {
			logger.FromContext(ctx).Warn("meili search failed; using DB fallback", slog.Any("err", err))
		}
	}
	return s.dbSearch(ctx, query, locale, limit)
}

func (s *SearchService) dbSearch(ctx context.Context, query, locale string, limit int) ([]SearchResult, error) {
	products, err := s.catalog.SearchProducts(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, len(products))
	for i := range products {
		ids[i] = products[i].ID
	}
	tr, _ := s.translations.LoadForEntities(ctx, models.EntityProduct, ids)
	res := i18n.NewResolver(s.defaultLocale, tr)

	out := make([]SearchResult, 0, len(products))
	for i := range products {
		p := &products[i]
		out = append(out, SearchResult{
			ID:        p.ID,
			Slug:      p.Slug,
			Name:      res.Field(models.EntityProduct, p.ID, models.FieldName, locale),
			PriceFrom: priceFrom(p),
			ImageURL:  s.resolveURL(p.ImageURL),
			Source:    "db",
		})
	}
	return out, nil
}

// priceFrom returns the lowest price to advertise (min variant or base price).
func priceFrom(p *models.Product) int64 {
	if p.PricingMode == models.PricingUnit && len(p.Variants) > 0 {
		min := p.Variants[0].PriceCents
		for _, v := range p.Variants[1:] {
			if v.PriceCents < min {
				min = v.PriceCents
			}
		}
		return min
	}
	return p.BasePriceCents
}

func (s *SearchService) mapMeiliHits(hits []search.SearchHit, locale string) []SearchResult {
	out := make([]SearchResult, 0, len(hits))
	for _, h := range hits {
		r := SearchResult{Source: "meili"}
		if v, ok := h["id"].(float64); ok {
			r.ID = uint64(v)
		}
		r.Slug, _ = h["slug"].(string)
		img, _ := h["image_url"].(string)
		r.ImageURL = s.resolveURL(img)
		if v, ok := h["price_from"].(float64); ok {
			r.PriceFrom = int64(v)
		}
		if name, ok := h["name_"+locale].(string); ok && name != "" {
			r.Name = name
		} else if name, ok := h["name_en"].(string); ok {
			r.Name = name
		}
		out = append(out, r)
	}
	return out
}

func (s *SearchService) resolveURL(raw string) string {
	if raw == "" || strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return s.baseURL + raw
}

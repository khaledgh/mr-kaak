package services

import (
	"context"
	"errors"

	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/search"
)

// IndexProduct (re)indexes a single product into Meilisearch. Called by the
// search.index Asynq worker handler and by the admin reindex.
func (s *SearchService) IndexProduct(ctx context.Context, productID uint64) error {
	if s.meili == nil {
		return nil
	}
	p, err := s.catalog.FindProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return s.meili.DeleteDocument(ctx, productID) // gone -> de-index
		}
		return err
	}
	doc := s.buildDocument(ctx, p)
	return s.meili.UpsertDocuments(ctx, []search.Document{doc})
}

// DeleteProductFromIndex removes a product document.
func (s *SearchService) DeleteProductFromIndex(ctx context.Context, productID uint64) error {
	if s.meili == nil {
		return nil
	}
	return s.meili.DeleteDocument(ctx, productID)
}

// Reindex rebuilds the entire index from MySQL (full resync / recovery).
func (s *SearchService) Reindex(ctx context.Context) (int, error) {
	if s.meili == nil {
		return 0, ErrSearchUnavailable
	}
	if err := s.meili.EnsureIndex(ctx); err != nil {
		return 0, err
	}
	products, err := s.catalog.AllProductsForIndex(ctx)
	if err != nil {
		return 0, err
	}
	docs := make([]search.Document, 0, len(products))
	for i := range products {
		docs = append(docs, s.buildDocument(ctx, &products[i]))
	}
	if len(docs) == 0 {
		return 0, nil
	}
	if err := s.meili.UpsertDocuments(ctx, docs); err != nil {
		return 0, err
	}
	return len(docs), nil
}

// buildDocument assembles the multi-language search document for a product.
func (s *SearchService) buildDocument(ctx context.Context, p *models.Product) search.Document {
	tr, _ := s.translations.LoadForEntity(ctx, models.EntityProduct, p.ID)
	get := func(locale, field string) string {
		for _, t := range tr {
			if t.Locale == locale && t.Field == field {
				return t.Value
			}
		}
		return ""
	}
	return search.Document{
		ID:          p.ID,
		Slug:        p.Slug,
		NameEN:      get("en", models.FieldName),
		NameAR:      get("ar", models.FieldName),
		NameFR:      get("fr", models.FieldName),
		DescEN:      get("en", models.FieldDescription),
		DescAR:      get("ar", models.FieldDescription),
		DescFR:      get("fr", models.FieldDescription),
		PriceFrom:   priceFrom(p),
		IsAvailable: p.IsAvailable,
		ImageURL:    p.ImageURL,
	}
}

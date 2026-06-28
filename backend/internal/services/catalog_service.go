package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mrkaak/restaurant-api/internal/cache"
	"github.com/mrkaak/restaurant-api/internal/i18n"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
)

// SearchIndexer is the seam through which the catalog keeps the search index in
// sync. The Asynq enqueuer implements it; calls are best-effort (no error) so a
// Redis/queue outage never blocks a catalog edit (plan §3.5).
type SearchIndexer interface {
	IndexProduct(ctx context.Context, productID uint64)
	DeleteProduct(ctx context.Context, productID uint64)
}

// noopIndexer is used when no indexer is wired (e.g. tests).
type noopIndexer struct{}

func (noopIndexer) IndexProduct(context.Context, uint64)  {}
func (noopIndexer) DeleteProduct(context.Context, uint64) {}

// CatalogService assembles the public menu (cached, translated) and provides
// admin CRUD for categories/products. Every mutation invalidates the menu
// cache by bumping its version namespace (plan §3.4) and re-syncs search.
type CatalogService struct {
	catalog       *repository.CatalogRepo
	translations  *repository.TranslationRepo
	cache         *cache.Cache
	indexer       SearchIndexer
	defaultLocale string
	baseURL       string
}

func NewCatalogService(catalog *repository.CatalogRepo, tr *repository.TranslationRepo, c *cache.Cache, indexer SearchIndexer, defaultLocale, baseURL string) *CatalogService {
	if indexer == nil {
		indexer = noopIndexer{}
	}
	return &CatalogService{catalog: catalog, translations: tr, cache: c, indexer: indexer, defaultLocale: defaultLocale, baseURL: strings.TrimRight(baseURL, "/")}
}

// GetMenu returns the active menu for a locale, served from cache when warm.
// On a miss (or with Redis down) it loads from the DB and resolves translations.
func (s *CatalogService) GetMenu(ctx context.Context, locale string) ([]models.Category, error) {
	locale = s.normalizeLocale(locale)
	key := s.cache.MenuKey(ctx, locale)
	return cache.GetOrSet(ctx, s.cache, key, time.Duration(cache.TTLMenu)*time.Second,
		func(ctx context.Context) ([]models.Category, error) {
			return s.loadMenu(ctx, locale)
		})
}

func (s *CatalogService) loadMenu(ctx context.Context, locale string) ([]models.Category, error) {
	cats, err := s.catalog.LoadActiveMenu(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.translateCategories(ctx, cats, locale); err != nil {
		return nil, err
	}
	for i := range cats {
		cats[i].ImageURL = s.resolveURL(cats[i].ImageURL)
		for j := range cats[i].Products {
			cats[i].Products[j].ImageURL = s.resolveURL(cats[i].Products[j].ImageURL)
		}
	}
	return cats, nil
}

// GetProductBySlug returns one product (translated) for a locale.
func (s *CatalogService) GetProductBySlug(ctx context.Context, slug, locale string) (*models.Product, error) {
	locale = s.normalizeLocale(locale)
	p, err := s.catalog.FindProductBySlug(ctx, slug)
	if err != nil {
		return nil, mapNotFound(err)
	}
	tr, err := s.translations.LoadForEntity(ctx, models.EntityProduct, p.ID)
	if err != nil {
		return nil, err
	}
	res := i18n.NewResolver(s.defaultLocale, tr)
	applyProductText(res, p, locale)
	p.ImageURL = s.resolveURL(p.ImageURL)
	return p, nil
}

// ListCategories returns categories (active-only for public, all for admin),
// translated.
func (s *CatalogService) ListCategories(ctx context.Context, locale string, activeOnly bool) ([]models.Category, error) {
	locale = s.normalizeLocale(locale)
	cats, err := s.catalog.ListCategories(ctx, activeOnly)
	if err != nil {
		return nil, err
	}
	ids := categoryIDs(cats)
	tr, err := s.translations.LoadForEntities(ctx, models.EntityCategory, ids)
	if err != nil {
		return nil, err
	}
	res := i18n.NewResolver(s.defaultLocale, tr)
	for i := range cats {
		applyCategoryText(res, &cats[i], locale)
		cats[i].ImageURL = s.resolveURL(cats[i].ImageURL)
	}
	return cats, nil
}

// translateCategories loads category + product translations for the whole menu
// in two queries and fills the resolved Name/Description fields.
func (s *CatalogService) translateCategories(ctx context.Context, cats []models.Category, locale string) error {
	catIDs := categoryIDs(cats)
	var prodIDs []uint64
	for i := range cats {
		for j := range cats[i].Products {
			prodIDs = append(prodIDs, cats[i].Products[j].ID)
		}
	}

	catTr, err := s.translations.LoadForEntities(ctx, models.EntityCategory, catIDs)
	if err != nil {
		return err
	}
	prodTr, err := s.translations.LoadForEntities(ctx, models.EntityProduct, prodIDs)
	if err != nil {
		return err
	}
	catRes := i18n.NewResolver(s.defaultLocale, catTr)
	prodRes := i18n.NewResolver(s.defaultLocale, prodTr)

	for i := range cats {
		applyCategoryText(catRes, &cats[i], locale)
		for j := range cats[i].Products {
			applyProductText(prodRes, &cats[i].Products[j], locale)
		}
	}
	return nil
}

// AdminListCategories returns all categories with every locale's translations
// attached — used by the admin panel list/edit UI.
func (s *CatalogService) AdminListCategories(ctx context.Context) ([]AdminCategoryDTO, error) {
	cats, err := s.catalog.ListCategories(ctx, false)
	if err != nil {
		return nil, err
	}
	ids := categoryIDs(cats)
	rows, err := s.translations.LoadForEntities(ctx, models.EntityCategory, ids)
	if err != nil {
		return nil, err
	}
	trMap := groupTranslations(rows)
	out := make([]AdminCategoryDTO, len(cats))
	for i, c := range cats {
		out[i] = AdminCategoryDTO{
			ID: c.ID, Slug: c.Slug, SortOrder: c.SortOrder,
			ImageURL: s.resolveURL(c.ImageURL), IsActive: c.IsActive,
			Translations: trMap[c.ID],
		}
	}
	return out, nil
}

// AdminListProducts returns all products (optionally filtered by category) with
// every locale's translations attached.
func (s *CatalogService) AdminListProducts(ctx context.Context, categoryID uint64) ([]AdminProductDTO, error) {
	products, err := s.catalog.ListProducts(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}
	rows, err := s.translations.LoadForEntities(ctx, models.EntityProduct, ids)
	if err != nil {
		return nil, err
	}
	trMap := groupTranslations(rows)
	out := make([]AdminProductDTO, len(products))
	for i, p := range products {
		allergens := []string(p.Allergens)
		if allergens == nil {
			allergens = []string{}
		}
		out[i] = AdminProductDTO{
			ID: p.ID, CategoryID: p.CategoryID, Slug: p.Slug,
			PricingMode: string(p.PricingMode), BasePriceCents: p.BasePriceCents,
			IsPreorder: p.IsPreorder, PreorderLeadHours: p.PreorderLeadHours,
			IsAvailable: p.IsAvailable, ImageURL: s.resolveURL(p.ImageURL),
			Allergens: allergens, SortOrder: p.SortOrder,
			Translations:   trMap[p.ID],
			Variants:       p.Variants,
			ModifierGroups: p.ModifierGroups,
		}
	}
	return out, nil
}

// groupTranslations pivots a flat slice of translation rows into a map of
// entity_id → locale → {name, description}.
func groupTranslations(rows []models.Translation) map[uint64]map[string]LocalizedText {
	out := make(map[uint64]map[string]LocalizedText)
	for _, t := range rows {
		if out[t.EntityID] == nil {
			out[t.EntityID] = make(map[string]LocalizedText)
		}
		lt := out[t.EntityID][t.Locale]
		switch t.Field {
		case models.FieldName:
			lt.Name = t.Value
		case models.FieldDescription:
			lt.Description = t.Value
		}
		out[t.EntityID][t.Locale] = lt
	}
	return out
}

// --- Admin: categories ---

func (s *CatalogService) CreateCategory(ctx context.Context, in CategoryInput) (*models.Category, error) {
	if _, ok := in.Translations[s.defaultLocale]; !ok {
		return nil, ErrDefaultLocaleRequired
	}
	c := &models.Category{
		Slug:      strings.ToLower(strings.TrimSpace(in.Slug)),
		SortOrder: in.SortOrder,
		ImageURL:  toRelativePath(in.ImageURL),
		IsActive:  derefBool(in.IsActive, true),
	}
	if err := s.catalog.CreateCategory(ctx, c); err != nil {
		if repository.IsDuplicateKey(err) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	if err := s.saveTranslations(ctx, models.EntityCategory, c.ID, in.Translations); err != nil {
		return nil, err
	}
	s.invalidateMenu(ctx)
	return s.hydrateCategory(ctx, c, s.defaultLocale)
}

func (s *CatalogService) UpdateCategory(ctx context.Context, id uint64, in CategoryInput) (*models.Category, error) {
	c, err := s.catalog.FindCategory(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	c.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	c.SortOrder = in.SortOrder
	c.ImageURL = toRelativePath(in.ImageURL)
	c.IsActive = derefBool(in.IsActive, c.IsActive)
	if err := s.catalog.UpdateCategory(ctx, c); err != nil {
		if repository.IsDuplicateKey(err) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	if err := s.saveTranslations(ctx, models.EntityCategory, c.ID, in.Translations); err != nil {
		return nil, err
	}
	s.invalidateMenu(ctx)
	return s.hydrateCategory(ctx, c, s.defaultLocale)
}

// SetCategoryActive sets the is_active flag on a category and returns the
// updated category with the default-locale translation applied.
func (s *CatalogService) SetCategoryActive(ctx context.Context, id uint64, active bool) (*models.Category, error) {
	cat, err := s.catalog.FindCategory(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	cat.IsActive = active
	if err := s.catalog.UpdateCategory(ctx, cat); err != nil {
		return nil, err
	}
	return s.hydrateCategory(ctx, cat, s.defaultLocale)
}

func (s *CatalogService) DeleteCategory(ctx context.Context, id uint64) error {
	hasProducts, err := s.catalog.CategoryHasProducts(ctx, id)
	if err != nil {
		return err
	}
	if hasProducts {
		return ErrCategoryNotEmpty
	}
	if err := s.catalog.DeleteCategory(ctx, id); err != nil {
		return err
	}
	_ = s.translations.DeleteForEntity(ctx, models.EntityCategory, id)
	s.invalidateMenu(ctx)
	return nil
}

// --- Admin: products ---

func (s *CatalogService) CreateProduct(ctx context.Context, in ProductInput) (*models.Product, error) {
	if _, ok := in.Translations[s.defaultLocale]; !ok {
		return nil, ErrDefaultLocaleRequired
	}
	if _, err := s.catalog.FindCategory(ctx, in.CategoryID); err != nil {
		return nil, ErrNotFound // category must exist
	}

	p := &models.Product{
		CategoryID:        in.CategoryID,
		Slug:              strings.ToLower(strings.TrimSpace(in.Slug)),
		PricingMode:       models.PricingMode(in.PricingMode),
		BasePriceCents:    in.BasePriceCents,
		IsPreorder:        in.IsPreorder,
		PreorderLeadHours: in.PreorderLeadHours,
		IsAvailable:       derefBool(in.IsAvailable, true),
		ImageURL:          toRelativePath(in.ImageURL),
		Allergens:         models.Allergens(in.Allergens),
		SortOrder:         in.SortOrder,
	}
	if err := s.catalog.CreateProduct(ctx, p); err != nil {
		if repository.IsDuplicateKey(err) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	if err := s.saveTranslations(ctx, models.EntityProduct, p.ID, in.Translations); err != nil {
		return nil, err
	}
	if err := s.catalog.ReplaceVariants(ctx, p.ID, toVariants(in.Variants)); err != nil {
		return nil, err
	}
	if err := s.catalog.ReplaceModifierGroups(ctx, p.ID, toModifierGroups(in.ModifierGroups)); err != nil {
		return nil, err
	}
	s.invalidateMenu(ctx)
	s.indexer.IndexProduct(ctx, p.ID)
	return s.GetProductBySlug(ctx, p.Slug, s.defaultLocale)
}

func (s *CatalogService) UpdateProduct(ctx context.Context, id uint64, in ProductInput) (*models.Product, error) {
	p, err := s.catalog.FindProductByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	p.CategoryID = in.CategoryID
	p.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	p.PricingMode = models.PricingMode(in.PricingMode)
	p.BasePriceCents = in.BasePriceCents
	p.IsPreorder = in.IsPreorder
	p.PreorderLeadHours = in.PreorderLeadHours
	p.IsAvailable = derefBool(in.IsAvailable, p.IsAvailable)
	p.ImageURL = toRelativePath(in.ImageURL)
	p.Allergens = models.Allergens(in.Allergens)
	p.SortOrder = in.SortOrder
	if err := s.catalog.UpdateProduct(ctx, p); err != nil {
		if repository.IsDuplicateKey(err) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	if err := s.saveTranslations(ctx, models.EntityProduct, p.ID, in.Translations); err != nil {
		return nil, err
	}
	if err := s.catalog.ReplaceVariants(ctx, p.ID, toVariants(in.Variants)); err != nil {
		return nil, err
	}
	if err := s.catalog.ReplaceModifierGroups(ctx, p.ID, toModifierGroups(in.ModifierGroups)); err != nil {
		return nil, err
	}
	s.invalidateMenu(ctx)
	s.indexer.IndexProduct(ctx, p.ID)
	return s.GetProductBySlug(ctx, p.Slug, s.defaultLocale)
}

func (s *CatalogService) SetProductAvailability(ctx context.Context, id uint64, available bool) error {
	if err := s.catalog.UpdateProductFields(ctx, id, map[string]any{"is_available": available}); err != nil {
		return err
	}
	s.invalidateMenu(ctx)
	s.indexer.IndexProduct(ctx, id)
	return nil
}

func (s *CatalogService) DeleteProduct(ctx context.Context, id uint64) error {
	if err := s.catalog.DeleteProduct(ctx, id); err != nil {
		return err
	}
	_ = s.translations.DeleteForEntity(ctx, models.EntityProduct, id)
	s.invalidateMenu(ctx)
	s.indexer.DeleteProduct(ctx, id)
	return nil
}

// FlushCache clears the whole cache DB (admin "clear cache" button, §3.4).
func (s *CatalogService) FlushCache(ctx context.Context) error {
	return s.cache.FlushAll(ctx)
}

// --- helpers ---

func (s *CatalogService) invalidateMenu(ctx context.Context) { s.cache.BumpMenuVersion(ctx) }

func (s *CatalogService) saveTranslations(ctx context.Context, entityType string, id uint64, tr map[string]LocalizedText) error {
	rows := make([]models.Translation, 0, len(tr)*2)
	for locale, txt := range tr {
		locale = s.normalizeLocale(locale)
		rows = append(rows, models.Translation{
			EntityType: entityType, EntityID: id, Locale: locale,
			Field: models.FieldName, Value: strings.TrimSpace(txt.Name),
		})
		if strings.TrimSpace(txt.Description) != "" {
			rows = append(rows, models.Translation{
				EntityType: entityType, EntityID: id, Locale: locale,
				Field: models.FieldDescription, Value: strings.TrimSpace(txt.Description),
			})
		}
	}
	return s.translations.UpsertMany(ctx, rows)
}

func (s *CatalogService) hydrateCategory(ctx context.Context, c *models.Category, locale string) (*models.Category, error) {
	tr, err := s.translations.LoadForEntity(ctx, models.EntityCategory, c.ID)
	if err != nil {
		return nil, err
	}
	applyCategoryText(i18n.NewResolver(s.defaultLocale, tr), c, locale)
	c.ImageURL = s.resolveURL(c.ImageURL)
	return c, nil
}

func (s *CatalogService) normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		return s.defaultLocale
	}
	return locale
}

func applyCategoryText(res *i18n.Resolver, c *models.Category, locale string) {
	c.Name = res.Field(models.EntityCategory, c.ID, models.FieldName, locale)
	c.Description = res.Field(models.EntityCategory, c.ID, models.FieldDescription, locale)
}

func applyProductText(res *i18n.Resolver, p *models.Product, locale string) {
	p.Name = res.Field(models.EntityProduct, p.ID, models.FieldName, locale)
	p.Description = res.Field(models.EntityProduct, p.ID, models.FieldDescription, locale)
}

func categoryIDs(cats []models.Category) []uint64 {
	ids := make([]uint64, len(cats))
	for i := range cats {
		ids[i] = cats[i].ID
	}
	return ids
}

func toVariants(in []VariantInput) []models.ProductVariant {
	out := make([]models.ProductVariant, len(in))
	for i, v := range in {
		out[i] = models.ProductVariant{
			SKU: v.SKU, Label: v.Label, PriceCents: v.PriceCents,
			IsDefault: v.IsDefault, SortOrder: v.SortOrder,
		}
	}
	return out
}

func toModifierGroups(in []ModifierGroupInput) []models.ModifierGroup {
	out := make([]models.ModifierGroup, len(in))
	for i, g := range in {
		mods := make([]models.Modifier, len(g.Modifiers))
		for j, m := range g.Modifiers {
			mods[j] = models.Modifier{
				Label: m.Label, PriceDeltaCents: m.PriceDeltaCents,
				IsDefault: m.IsDefault, IsAvailable: derefBool(m.IsAvailable, true), SortOrder: m.SortOrder,
			}
		}
		out[i] = models.ModifierGroup{
			Label: g.Label, MinSelect: g.MinSelect, MaxSelect: g.MaxSelect,
			IsRequired: g.IsRequired, SortOrder: g.SortOrder, Modifiers: mods,
		}
	}
	return out
}

func derefBool(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func mapNotFound(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// resolveURL prepends the public base URL to relative paths.
func (s *CatalogService) resolveURL(raw string) string {
	if raw == "" || strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return s.baseURL + raw
}

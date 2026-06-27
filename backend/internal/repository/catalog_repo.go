package repository

import (
	"context"

	"github.com/mrkaak/restaurant-api/internal/models"
	"gorm.io/gorm"
)

// CatalogRepo is the data-access layer for categories, products, variants, and
// modifiers.
type CatalogRepo struct {
	db *gorm.DB
}

func NewCatalogRepo(db *gorm.DB) *CatalogRepo { return &CatalogRepo{db: db} }

// orderedModifiers preloads modifiers within their sort order.
func orderedModifiers(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }

// LoadActiveMenu loads all active categories with their available products and
// each product's variants and modifier groups/modifiers, fully ordered. This
// is the single query path behind the cached public menu.
func (r *CatalogRepo) LoadActiveMenu(ctx context.Context) ([]models.Category, error) {
	var cats []models.Category
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("Products", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_available = ?", true).Order("sort_order ASC, id ASC")
		}).
		Preload("Products.Variants", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Products.ModifierGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Products.ModifierGroups.Modifiers", orderedModifiers).
		Order("sort_order ASC, id ASC").
		Find(&cats).Error
	return cats, err
}

// FindProductBySlug loads a single product with its relations.
func (r *CatalogRepo) FindProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).
		Preload("Variants", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("ModifierGroups", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("ModifierGroups.Modifiers", orderedModifiers).
		Where("slug = ?", slug).First(&p).Error
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &p, nil
}

// FindProductByID loads a product (with relations) by id — used at checkout to
// re-price from the source of truth.
func (r *CatalogRepo) FindProductByID(ctx context.Context, id uint64) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).
		Preload("Variants").
		Preload("ModifierGroups").
		Preload("ModifierGroups.Modifiers").
		First(&p, id).Error
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &p, nil
}

// SearchProducts is the DB fallback for search (used when Meilisearch is
// unavailable). It matches the query against any locale's product name in the
// translations table and returns the available products with relations.
func (r *CatalogRepo) SearchProducts(ctx context.Context, query string, limit int) ([]models.Product, error) {
	like := "%" + query + "%"
	var ids []uint64
	err := r.db.WithContext(ctx).
		Model(&models.Translation{}).
		Distinct("entity_id").
		Where("entity_type = ? AND field = ? AND value LIKE ?", models.EntityProduct, models.FieldName, like).
		Limit(limit).Pluck("entity_id", &ids).Error
	if err != nil {
		return nil, err
	}
	// Also match by slug for ascii queries.
	var slugIDs []uint64
	_ = r.db.WithContext(ctx).Model(&models.Product{}).
		Where("slug LIKE ?", like).Limit(limit).Pluck("id", &slugIDs).Error
	ids = append(ids, slugIDs...)
	if len(ids) == 0 {
		return nil, nil
	}

	var products []models.Product
	err = r.db.WithContext(ctx).
		Preload("Variants").Preload("ModifierGroups").Preload("ModifierGroups.Modifiers").
		Where("id IN ? AND is_available = ?", ids, true).
		Limit(limit).Find(&products).Error
	return products, err
}

// AllProductsForIndex loads every product with relations + category, for a full
// Meilisearch reindex.
func (r *CatalogRepo) AllProductsForIndex(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).Preload("Variants").Find(&products).Error
	return products, err
}

// --- Categories ---

func (r *CatalogRepo) ListCategories(ctx context.Context, activeOnly bool) ([]models.Category, error) {
	q := r.db.WithContext(ctx).Order("sort_order ASC, id ASC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var cats []models.Category
	return cats, q.Find(&cats).Error
}

// ListProducts returns all products (with variants + modifier groups), optionally
// filtered by category. Used by the admin list endpoint.
func (r *CatalogRepo) ListProducts(ctx context.Context, categoryID uint64) ([]models.Product, error) {
	q := r.db.WithContext(ctx).
		Preload("Variants", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("ModifierGroups", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC, id ASC") }).
		Preload("ModifierGroups.Modifiers", orderedModifiers).
		Order("sort_order ASC, id ASC")
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	var products []models.Product
	return products, q.Find(&products).Error
}

func (r *CatalogRepo) CreateCategory(ctx context.Context, c *models.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CatalogRepo) FindCategory(ctx context.Context, id uint64) (*models.Category, error) {
	var c models.Category
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &c, nil
}

func (r *CatalogRepo) UpdateCategory(ctx context.Context, c *models.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CatalogRepo) DeleteCategory(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, id).Error
}

// CategoryHasProducts reports whether a category still has products (used to
// block deletion of a non-empty category).
func (r *CatalogRepo) CategoryHasProducts(ctx context.Context, id uint64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("category_id = ?", id).Count(&n).Error
	return n > 0, err
}

// --- Products ---

func (r *CatalogRepo) CreateProduct(ctx context.Context, p *models.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *CatalogRepo) UpdateProduct(ctx context.Context, p *models.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *CatalogRepo) UpdateProductFields(ctx context.Context, id uint64, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.Product{Base: models.Base{ID: id}}).Updates(fields).Error
}

func (r *CatalogRepo) DeleteProduct(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Variants", "ModifierGroups").Delete(&models.Product{Base: models.Base{ID: id}}).Error
}

// --- Variants & modifiers (replace-on-save children) ---

func (r *CatalogRepo) ReplaceVariants(ctx context.Context, productID uint64, variants []models.ProductVariant) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", productID).Delete(&models.ProductVariant{}).Error; err != nil {
			return err
		}
		if len(variants) == 0 {
			return nil
		}
		for i := range variants {
			variants[i].ProductID = productID
		}
		return tx.Create(&variants).Error
	})
}

func (r *CatalogRepo) ReplaceModifierGroups(ctx context.Context, productID uint64, groups []models.ModifierGroup) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Cascade delete removes child modifiers via FK ON DELETE CASCADE only
		// for hard deletes; here we explicitly clear both for soft-delete safety.
		var ids []uint64
		if err := tx.Model(&models.ModifierGroup{}).Where("product_id = ?", productID).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) > 0 {
			if err := tx.Where("group_id IN ?", ids).Delete(&models.Modifier{}).Error; err != nil {
				return err
			}
			if err := tx.Where("product_id = ?", productID).Delete(&models.ModifierGroup{}).Error; err != nil {
				return err
			}
		}
		for i := range groups {
			groups[i].ProductID = productID
			groups[i].ID = 0
		}
		if len(groups) > 0 {
			return tx.Create(&groups).Error
		}
		return nil
	})
}

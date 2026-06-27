package services

import (
	"context"
	"strings"

	"github.com/mrkaak/restaurant-api/internal/geo"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
)

// ZoneService matches a delivery point against zones and computes the fee, and
// provides admin CRUD. Matching follows plan §5.2: global zone first, then
// tighter per-product zones.
type ZoneService struct {
	zones *repository.ZoneRepo
}

func NewZoneService(zones *repository.ZoneRepo) *ZoneService {
	return &ZoneService{zones: zones}
}

// DeliveryQuote is the result of matching a point against the zones.
type DeliveryQuote struct {
	Deliverable             bool     `json:"deliverable"`
	FeeCents                int64    `json:"fee_cents"`
	MinOrderCents           int64    `json:"min_order_cents"`
	MatchedZoneID           uint64   `json:"matched_zone_id,omitempty"`
	MatchedZoneName         string   `json:"matched_zone_name,omitempty"`
	Reason                  string   `json:"reason,omitempty"`
	UndeliverableProductIDs []uint64 `json:"undeliverable_product_ids,omitempty"`
}

// Quote determines whether a point is deliverable and at what fee.
func (s *ZoneService) Quote(ctx context.Context, p geo.Point, productIDs []uint64) (*DeliveryQuote, error) {
	// 1. Global zone: the store must deliver to this point at all.
	globals, err := s.zones.ActiveGlobal(ctx)
	if err != nil {
		return nil, err
	}
	var matched *models.DeliveryZone
	for i := range globals {
		if globals[i].Contains(p) {
			matched = &globals[i]
			break
		}
	}
	if matched == nil {
		return &DeliveryQuote{Deliverable: false, Reason: "we don't deliver to your area yet"}, nil
	}

	// 2. Per-product zones: a product with its own zones must be reachable.
	productZones, err := s.zones.ActiveForProducts(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	byProduct := map[uint64][]*models.DeliveryZone{}
	for i := range productZones {
		if pid := productZones[i].ProductID; pid != nil {
			byProduct[*pid] = append(byProduct[*pid], &productZones[i])
		}
	}
	var undeliverable []uint64
	for pid, zs := range byProduct {
		ok := false
		for _, z := range zs {
			if z.Contains(p) {
				ok = true
				break
			}
		}
		if !ok {
			undeliverable = append(undeliverable, pid)
		}
	}
	if len(undeliverable) > 0 {
		return &DeliveryQuote{
			Deliverable:             false,
			Reason:                  "some items can't be delivered to this address",
			UndeliverableProductIDs: undeliverable,
		}, nil
	}

	return &DeliveryQuote{
		Deliverable:     true,
		FeeCents:        matched.FeeCents,
		MinOrderCents:   matched.MinOrderCents,
		MatchedZoneID:   matched.ID,
		MatchedZoneName: matched.Name,
	}, nil
}

// --- Admin CRUD ---

func (s *ZoneService) List(ctx context.Context) ([]models.DeliveryZone, error) {
	return s.zones.List(ctx, false)
}

func (s *ZoneService) Create(ctx context.Context, in ZoneInput) (*models.DeliveryZone, error) {
	z, err := in.toModel()
	if err != nil {
		return nil, err
	}
	if err := s.zones.Create(ctx, z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *ZoneService) Update(ctx context.Context, id uint64, in ZoneInput) (*models.DeliveryZone, error) {
	existing, err := s.zones.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	updated, err := in.toModel()
	if err != nil {
		return nil, err
	}
	updated.ID = existing.ID
	if err := s.zones.Update(ctx, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *ZoneService) Delete(ctx context.Context, id uint64) error {
	return s.zones.Delete(ctx, id)
}

func (in ZoneInput) toModel() (*models.DeliveryZone, error) {
	z := &models.DeliveryZone{
		Name:          strings.TrimSpace(in.Name),
		Scope:         models.ZoneScope(in.Scope),
		ProductID:     in.ProductID,
		Shape:         models.ZoneShape(in.Shape),
		FeeCents:      in.FeeCents,
		MinOrderCents: in.MinOrderCents,
		IsActive:      derefBool(in.IsActive, true),
		SortOrder:     in.SortOrder,
	}
	switch z.Shape {
	case models.ShapeRadius:
		if in.CenterLat == nil || in.CenterLng == nil || in.RadiusKm == nil || *in.RadiusKm <= 0 {
			return nil, ErrInvalidZone
		}
		z.CenterLat, z.CenterLng, z.RadiusKm = in.CenterLat, in.CenterLng, in.RadiusKm
	case models.ShapePolygon:
		if len(in.PolygonGeoJSON) < 3 {
			return nil, ErrInvalidZone
		}
		z.PolygonGeoJSON = models.GeoPolygon(in.PolygonGeoJSON)
	default:
		return nil, ErrInvalidZone
	}
	if z.Scope == models.ScopeProduct && z.ProductID == nil {
		return nil, ErrInvalidZone
	}
	return z, nil
}

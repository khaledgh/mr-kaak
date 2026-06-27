package services

// ZoneInput is the admin create/update payload for a delivery zone.
type ZoneInput struct {
	Name           string       `json:"name" validate:"required,max=120"`
	Scope          string       `json:"scope" validate:"required,oneof=global product"`
	ProductID      *uint64      `json:"product_id"`
	Shape          string       `json:"shape" validate:"required,oneof=radius polygon"`
	CenterLat      *float64     `json:"center_lat" validate:"omitempty,latitude"`
	CenterLng      *float64     `json:"center_lng" validate:"omitempty,longitude"`
	RadiusKm       *float64     `json:"radius_km" validate:"omitempty,gt=0"`
	PolygonGeoJSON [][2]float64 `json:"polygon_geojson"`
	FeeCents       int64        `json:"fee_cents" validate:"min=0"`
	MinOrderCents  int64        `json:"min_order_cents" validate:"min=0"`
	IsActive       *bool        `json:"is_active"`
	SortOrder      int          `json:"sort_order" validate:"omitempty,min=0"`
}

// QuoteInput asks whether a point is deliverable for a set of products.
type QuoteInput struct {
	Lat        float64  `json:"lat" validate:"required,latitude"`
	Lng        float64  `json:"lng" validate:"required,longitude"`
	ProductIDs []uint64 `json:"product_ids"`
}

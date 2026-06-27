package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/mrkaak/restaurant-api/internal/geo"
)

type ZoneScope string

const (
	ScopeGlobal  ZoneScope = "global"
	ScopeProduct ZoneScope = "product"
)

type ZoneShape string

const (
	ShapeRadius  ZoneShape = "radius"
	ShapePolygon ZoneShape = "polygon"
)

// GeoPolygon is a ring of [lng, lat] pairs stored as JSON (GeoJSON-style
// coordinate order: longitude first).
type GeoPolygon [][2]float64

func (g GeoPolygon) Value() (driver.Value, error) {
	if g == nil {
		return nil, nil
	}
	return json.Marshal(g)
}

func (g *GeoPolygon) Scan(src any) error {
	if src == nil {
		*g = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, g)
	case string:
		return json.Unmarshal([]byte(v), g)
	default:
		return errors.New("geopolygon: unsupported scan type")
	}
}

// Points converts the stored [lng,lat] ring into geo.Point values.
func (g GeoPolygon) Points() []geo.Point {
	pts := make([]geo.Point, len(g))
	for i, c := range g {
		pts[i] = geo.Point{Lng: c[0], Lat: c[1]}
	}
	return pts
}

// DeliveryZone is a deliverable area with a fee and minimum order.
type DeliveryZone struct {
	Base
	Name           string     `gorm:"size:120;not null" json:"name"`
	Scope          ZoneScope  `gorm:"type:enum('global','product');not null;default:global" json:"scope"`
	ProductID      *uint64    `gorm:"index" json:"product_id,omitempty"`
	Shape          ZoneShape  `gorm:"type:enum('radius','polygon');not null;default:radius" json:"shape"`
	CenterLat      *float64   `gorm:"type:decimal(10,7)" json:"center_lat,omitempty"`
	CenterLng      *float64   `gorm:"type:decimal(10,7)" json:"center_lng,omitempty"`
	RadiusKm       *float64   `gorm:"type:decimal(8,3)" json:"radius_km,omitempty"`
	PolygonGeoJSON GeoPolygon `gorm:"column:polygon_geojson;type:json" json:"polygon_geojson,omitempty"`
	FeeCents       int64      `gorm:"not null;default:0" json:"fee_cents"`
	MinOrderCents  int64      `gorm:"not null;default:0" json:"min_order_cents"`
	IsActive       bool       `gorm:"not null;default:true" json:"is_active"`
	SortOrder      int        `gorm:"not null;default:0" json:"sort_order"`
}

// Contains reports whether p is inside this zone (radius or polygon).
func (z *DeliveryZone) Contains(p geo.Point) bool {
	switch z.Shape {
	case ShapeRadius:
		if z.CenterLat == nil || z.CenterLng == nil || z.RadiusKm == nil {
			return false
		}
		return geo.WithinRadius(geo.Point{Lat: *z.CenterLat, Lng: *z.CenterLng}, p, *z.RadiusKm)
	case ShapePolygon:
		return geo.PointInPolygon(p, z.PolygonGeoJSON.Points())
	default:
		return false
	}
}

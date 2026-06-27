// Package geo implements the delivery-zone matching math (plan §5): great-circle
// distance for radius zones and ray-casting point-in-polygon for irregular
// boundaries. Kept dependency-free and pure so it is trivially testable and
// cache-friendly (zones change rarely).
package geo

import "math"

// Point is a WGS84 coordinate (degrees).
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

const earthRadiusKm = 6371.0088

// HaversineKm returns the great-circle distance between two points in km.
func HaversineKm(a, b Point) float64 {
	lat1 := rad(a.Lat)
	lat2 := rad(b.Lat)
	dLat := rad(b.Lat - a.Lat)
	dLng := rad(b.Lng - a.Lng)

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusKm * math.Asin(math.Min(1, math.Sqrt(h)))
}

// WithinRadius reports whether p is within radiusKm of center.
func WithinRadius(center, p Point, radiusKm float64) bool {
	return HaversineKm(center, p) <= radiusKm
}

// PointInPolygon reports whether p lies inside the polygon (ray casting).
// The polygon is an ordered ring of vertices; it need not be explicitly closed.
// Points exactly on an edge are treated as inside for delivery purposes.
func PointInPolygon(p Point, polygon []Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := polygon[i].Lng, polygon[i].Lat
		xj, yj := polygon[j].Lng, polygon[j].Lat

		intersect := ((yi > p.Lat) != (yj > p.Lat)) &&
			(p.Lng < (xj-xi)*(p.Lat-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

func rad(deg float64) float64 { return deg * math.Pi / 180 }

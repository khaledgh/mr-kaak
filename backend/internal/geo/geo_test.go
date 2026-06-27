package geo

import (
	"math"
	"testing"
)

func TestHaversineKm(t *testing.T) {
	// Vancouver downtown to Burnaby ~ 11-12 km.
	van := Point{Lat: 49.2827, Lng: -123.1207}
	bby := Point{Lat: 49.2488, Lng: -122.9805}
	d := HaversineKm(van, bby)
	if d < 9 || d > 14 {
		t.Errorf("distance = %.2f km, expected ~10-12", d)
	}
	// Zero distance.
	if got := HaversineKm(van, van); math.Abs(got) > 1e-9 {
		t.Errorf("self distance = %v, want 0", got)
	}
}

func TestWithinRadius(t *testing.T) {
	center := Point{Lat: 49.2827, Lng: -123.1207}
	near := Point{Lat: 49.29, Lng: -123.12} // ~0.8 km
	far := Point{Lat: 49.6, Lng: -123.5}    // tens of km
	if !WithinRadius(center, near, 5) {
		t.Error("near point should be within 5km")
	}
	if WithinRadius(center, far, 5) {
		t.Error("far point should be outside 5km")
	}
}

func TestPointInPolygon(t *testing.T) {
	// A unit square around the origin.
	square := []Point{{Lat: 0, Lng: 0}, {Lat: 0, Lng: 2}, {Lat: 2, Lng: 2}, {Lat: 2, Lng: 0}}
	if !PointInPolygon(Point{Lat: 1, Lng: 1}, square) {
		t.Error("center should be inside")
	}
	if PointInPolygon(Point{Lat: 3, Lng: 3}, square) {
		t.Error("outside point should be outside")
	}
	if PointInPolygon(Point{Lat: 1, Lng: 1}, square[:2]) {
		t.Error("degenerate polygon (<3 pts) should be outside")
	}
}

package pagination

import "testing"

func TestNewMeta(t *testing.T) {
	cases := []struct {
		name      string
		p         Params
		total     int64
		wantPages int
	}{
		{"exact", Params{Page: 1, PerPage: 20}, 40, 2},
		{"partial", Params{Page: 1, PerPage: 20}, 41, 3},
		{"empty", Params{Page: 1, PerPage: 20}, 0, 0},
		{"single", Params{Page: 1, PerPage: 20}, 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMeta(tc.p, tc.total)
			if m.TotalPages != tc.wantPages {
				t.Errorf("TotalPages = %d, want %d", m.TotalPages, tc.wantPages)
			}
		})
	}
}

func TestOffset(t *testing.T) {
	p := Params{Page: 3, PerPage: 25}
	if got := p.Offset(); got != 50 {
		t.Errorf("Offset() = %d, want 50", got)
	}
}

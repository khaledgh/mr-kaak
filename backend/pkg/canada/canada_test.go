package canada

import "testing"

func TestValidPostal(t *testing.T) {
	good := []string{"V6B 1A1", "v6b1a1", "K1A0B1", "M5V 2T6"}
	bad := []string{"99999", "D1A 1A1", "V6B 1A", "", "ZZZ ZZZ"}
	for _, s := range good {
		if !ValidPostal(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
	for _, s := range bad {
		if ValidPostal(s) {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}

func TestNormalizePostal(t *testing.T) {
	if got := NormalizePostal("v6b1a1"); got != "V6B 1A1" {
		t.Errorf("NormalizePostal = %q, want V6B 1A1", got)
	}
	if got := NormalizePostal("v6b 1a1"); got != "V6B 1A1" {
		t.Errorf("NormalizePostal = %q, want V6B 1A1", got)
	}
}

func TestValidProvince(t *testing.T) {
	if !ValidProvince("on") || !ValidProvince("QC") {
		t.Error("ON/QC should be valid")
	}
	if ValidProvince("ZZ") || ValidProvince("") {
		t.Error("ZZ/empty should be invalid")
	}
	if len(Provinces) != 13 {
		t.Errorf("expected 13 provinces/territories, got %d", len(Provinces))
	}
}

package validator

import "testing"

type addrDTO struct {
	Province string `json:"province_code" validate:"required,ca_province"`
	Postal   string `json:"postal_code" validate:"required,ca_postal"`
	Phone    string `json:"phone" validate:"omitempty,e164ca"`
	Role     string `json:"role" validate:"omitempty,role"`
}

func TestCustomRules(t *testing.T) {
	v := New()

	t.Run("valid", func(t *testing.T) {
		err := v.Struct(addrDTO{Province: "BC", Postal: "V6B 1A1", Phone: "604-555-0167", Role: "admin"})
		if err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("invalid collects fields", func(t *testing.T) {
		err := v.Struct(addrDTO{Province: "ZZ", Postal: "99999", Phone: "123", Role: "wizard"})
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Fatalf("expected *ValidationError, got %T", err)
		}
		if len(ve.Fields) != 4 {
			t.Errorf("expected 4 field errors, got %d: %v", len(ve.Fields), ve.Fields)
		}
		// Field names should be the JSON tags, not Go field names.
		for _, f := range ve.Fields {
			switch f.Field {
			case "province_code", "postal_code", "phone", "role":
			default:
				t.Errorf("unexpected field name %q", f.Field)
			}
		}
	})
}

func TestNormalizePhoneCA(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"604-555-0167", "+16045550167", true},
		{"(604) 555-0167", "+16045550167", true},
		{"+1 604 555 0167", "+16045550167", true},
		{"123", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, err := NormalizePhoneCA(tc.in)
		if tc.ok && (err != nil || got != tc.want) {
			t.Errorf("NormalizePhoneCA(%q) = %q,%v; want %q,nil", tc.in, got, err, tc.want)
		}
		if !tc.ok && err == nil {
			t.Errorf("NormalizePhoneCA(%q) expected error", tc.in)
		}
	}
}

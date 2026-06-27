// Package canada holds Canada-specific reference data and validation:
// province/territory codes (bilingual labels) and postal-code handling
// (plan §6). Kept provider-agnostic and dependency-free so it can be reused
// by validation, seeding, and the frontend bundle.
package canada

import (
	"regexp"
	"strings"
)

// Province is a province/territory with bilingual names.
type Province struct {
	Code   string `json:"code"`
	NameEN string `json:"name_en"`
	NameFR string `json:"name_fr"`
}

// Provinces lists all 13 Canadian provinces and territories.
var Provinces = []Province{
	{"AB", "Alberta", "Alberta"},
	{"BC", "British Columbia", "Colombie-Britannique"},
	{"MB", "Manitoba", "Manitoba"},
	{"NB", "New Brunswick", "Nouveau-Brunswick"},
	{"NL", "Newfoundland and Labrador", "Terre-Neuve-et-Labrador"},
	{"NS", "Nova Scotia", "Nouvelle-Écosse"},
	{"NT", "Northwest Territories", "Territoires du Nord-Ouest"},
	{"NU", "Nunavut", "Nunavut"},
	{"ON", "Ontario", "Ontario"},
	{"PE", "Prince Edward Island", "Île-du-Prince-Édouard"},
	{"QC", "Quebec", "Québec"},
	{"SK", "Saskatchewan", "Saskatchewan"},
	{"YT", "Yukon", "Yukon"},
}

var provinceSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Provinces))
	for _, p := range Provinces {
		m[p.Code] = struct{}{}
	}
	return m
}()

// ValidProvince reports whether code is a recognized province/territory code.
func ValidProvince(code string) bool {
	_, ok := provinceSet[strings.ToUpper(strings.TrimSpace(code))]
	return ok
}

// postalRe matches a Canadian postal code, excluding letters not used by
// Canada Post (D, F, I, O, Q, U; W/Z not allowed as first letter).
var postalRe = regexp.MustCompile(`^[ABCEGHJ-NPRSTVXY]\d[ABCEGHJ-NPRSTV-Z] ?\d[ABCEGHJ-NPRSTV-Z]\d$`)

// ValidPostal reports whether s is a valid Canadian postal code (with or
// without the interior space).
func ValidPostal(s string) bool {
	return postalRe.MatchString(strings.ToUpper(strings.TrimSpace(s)))
}

// NormalizePostal uppercases and inserts the canonical "A1A 1A1" space.
// Returns the input unchanged if it isn't a recognizable 6-char code.
func NormalizePostal(s string) string {
	up := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
	if len(up) != 6 {
		return strings.ToUpper(strings.TrimSpace(s))
	}
	return up[:3] + " " + up[3:]
}

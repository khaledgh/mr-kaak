package services

import "github.com/mrkaak/restaurant-api/internal/models"

// selectModifiers validates the chosen modifier ids against the product's
// modifier groups and returns the total price delta plus snapshots. It enforces
// that every chosen modifier belongs to the product and is available, respects
// each group's max_select, and that required groups have at least min_select.
func selectModifiers(p *models.Product, chosen []uint64) (int64, models.ModifierSnapshots, error) {
	// Index modifiers by id and remember each modifier's group.
	type modRef struct {
		mod   models.Modifier
		group *models.ModifierGroup
	}
	byID := map[uint64]modRef{}
	for gi := range p.ModifierGroups {
		g := &p.ModifierGroups[gi]
		for _, m := range g.Modifiers {
			byID[m.ID] = modRef{mod: m, group: g}
		}
	}

	chosenSet := map[uint64]bool{}
	perGroup := map[uint64]int{}
	var total int64
	snaps := make(models.ModifierSnapshots, 0, len(chosen))

	for _, id := range chosen {
		if chosenSet[id] {
			continue // ignore duplicates
		}
		ref, ok := byID[id]
		if !ok || !ref.mod.IsAvailable {
			return 0, nil, ErrModifierInvalid
		}
		chosenSet[id] = true
		perGroup[ref.group.ID]++
		total += ref.mod.PriceDeltaCents
		snaps = append(snaps, models.ModifierSnapshot{
			ModifierID: ref.mod.ID, Label: ref.mod.Label, PriceDeltaCents: ref.mod.PriceDeltaCents,
		})
	}

	// Enforce per-group selection bounds.
	for gi := range p.ModifierGroups {
		g := &p.ModifierGroups[gi]
		n := perGroup[g.ID]
		if g.MaxSelect > 0 && n > g.MaxSelect {
			return 0, nil, ErrModifierBounds
		}
		if g.IsRequired && n < g.MinSelect {
			return 0, nil, ErrModifierBounds
		}
	}
	return total, snaps, nil
}

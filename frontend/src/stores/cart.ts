import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { Modifier, Product, Variant } from "@/api/types";

// A cart line. Prices here are for display only; the server re-prices every
// line at checkout (never trust client prices — plan §9).
export interface CartLine {
  key: string; // product+variant+modifiers signature for de-dup
  productId: number;
  slug: string;
  name: string;
  imageUrl?: string;
  pricingMode: "unit" | "weight";
  variantId?: number;
  variantLabel?: string;
  modifierIds: number[];
  modifierLabels: string[];
  qty: number;
  weightGrams?: number;
  unitPriceCents: number; // display estimate
}

interface CartState {
  lines: CartLine[];
  add: (p: Product, opts: { variant?: Variant; modifiers?: Modifier[]; weightGrams?: number }) => void;
  setQty: (key: string, qty: number) => void;
  remove: (key: string) => void;
  clear: () => void;
  count: () => number;
  subtotalCents: () => number;
  productIds: () => number[];
}

function lineKey(productId: number, variantId?: number, modifierIds: number[] = []): string {
  return `${productId}:${variantId ?? 0}:${[...modifierIds].sort().join(",")}`;
}

export const useCart = create<CartState>()(
  persist(
    (set, get) => ({
      lines: [],
      add: (p, opts) => {
        const modifiers = opts.modifiers ?? [];
        const modifierIds = modifiers.map((m) => m.id);
        const key = lineKey(p.id, opts.variant?.id, modifierIds);
        const modSum = modifiers.reduce((s, m) => s + m.price_delta_cents, 0);
        const base = opts.variant ? opts.variant.price_cents : p.base_price_cents;
        const unit =
          p.pricing_mode === "weight" && opts.weightGrams
            ? Math.round((p.base_price_cents * opts.weightGrams) / 1000) + modSum
            : base + modSum;

        set((state) => {
          const existing = state.lines.find((l) => l.key === key);
          if (existing) {
            return {
              lines: state.lines.map((l) =>
                l.key === key ? { ...l, qty: l.qty + 1 } : l,
              ),
            };
          }
          const line: CartLine = {
            key,
            productId: p.id,
            slug: p.slug,
            name: p.name,
            imageUrl: p.image_url,
            pricingMode: p.pricing_mode,
            variantId: opts.variant?.id,
            variantLabel: opts.variant?.label,
            modifierIds,
            modifierLabels: modifiers.map((m) => m.label),
            qty: 1,
            weightGrams: opts.weightGrams,
            unitPriceCents: unit,
          };
          return { lines: [...state.lines, line] };
        });
      },
      setQty: (key, qty) =>
        set((state) => ({
          lines:
            qty <= 0
              ? state.lines.filter((l) => l.key !== key)
              : state.lines.map((l) => (l.key === key ? { ...l, qty } : l)),
        })),
      remove: (key) => set((state) => ({ lines: state.lines.filter((l) => l.key !== key) })),
      clear: () => set({ lines: [] }),
      count: () => get().lines.reduce((s, l) => s + l.qty, 0),
      subtotalCents: () => get().lines.reduce((s, l) => s + l.unitPriceCents * l.qty, 0),
      productIds: () => Array.from(new Set(get().lines.map((l) => l.productId))),
    }),
    { name: "cart" },
  ),
);

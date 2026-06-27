import { useState } from "react";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";
import type { Product, Variant } from "@/api/types";
import { useCart } from "@/stores/cart";
import { useToast } from "@/stores/toast";
import { formatMoney } from "@/lib/money";

export function ProductModal({ product, onClose }: { product: Product; onClose: () => void }) {
  const { t, i18n } = useTranslation();
  const add = useCart((s) => s.add);
  const toast = useToast((s) => s.show);

  const defaultVariant = product.variants?.find((v) => v.is_default) ?? product.variants?.[0];
  const [selected, setSelected] = useState<Variant | undefined>(defaultVariant);

  const price = selected ? selected.price_cents : product.base_price_cents;

  function handleAdd() {
    if (product.pricing_mode === "weight") {
      add(product, { weightGrams: 500 });
    } else {
      add(product, { variant: selected });
    }
    toast(t("product.added", "Added to cart!"));
    onClose();
  }

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="w-full max-w-lg animate-sheet rounded-t-3xl bg-white overflow-hidden shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Image header */}
        <div className="relative flex justify-center bg-gradient-to-br from-amber-50 to-brand-100 pt-6 pb-5">
          <button
            onClick={onClose}
            className="absolute right-4 top-4 flex h-8 w-8 items-center justify-center rounded-full bg-black/10 text-brand-800 hover:bg-black/20 transition text-lg leading-none"
          >
            ✕
          </button>
          {product.image_url ? (
            <img
              src={product.image_url}
              alt={product.name}
              className="h-40 w-40 rounded-full object-cover shadow-xl ring-4 ring-white"
            />
          ) : (
            <div className="flex h-40 w-40 items-center justify-center rounded-full bg-brand-200 text-7xl shadow-xl ring-4 ring-white">
              🍬
            </div>
          )}
          {product.is_preorder && (
            <span className="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-brand-600 px-3 py-1 text-xs font-medium text-white">
              {t("product.preorder")}
            </span>
          )}
        </div>

        {/* Content */}
        <div className="px-5 pt-4 pb-8">
          <h2 className="text-xl font-bold text-brand-900">{product.name}</h2>

          {product.description && (
            <p className="mt-2 text-sm leading-relaxed text-brand-700/70">{product.description}</p>
          )}

          {/* Variant picker */}
          {(product.variants ?? []).length > 1 && (
            <div className="mt-4">
              <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-brand-600">
                {t("product.size", "Size")}
              </p>
              <div className="flex flex-wrap gap-2">
                {product.variants!.map((v) => (
                  <button
                    key={v.id}
                    onClick={() => setSelected(v)}
                    className={`rounded-full border px-4 py-1.5 text-sm font-medium transition ${
                      selected?.id === v.id
                        ? "border-brand-600 bg-brand-600 text-white shadow-md"
                        : "border-brand-200 text-brand-700 hover:border-brand-400"
                    }`}
                  >
                    {v.label} — {formatMoney(v.price_cents, "CAD", i18n.language)}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Price + CTA */}
          <div className="mt-5 flex items-center justify-between gap-4">
            <div>
              <p className="text-xs uppercase tracking-wide text-brand-500">{t("product.price", "Price")}</p>
              <p className="text-2xl font-bold text-brand-700">
                {formatMoney(price, "CAD", i18n.language)}
                {product.pricing_mode === "weight" && (
                  <span className="text-sm font-normal"> {t("product.perKg")}</span>
                )}
              </p>
            </div>
            <button
              className="rounded-full bg-brand-600 px-8 py-3 font-semibold text-white shadow-lg transition hover:bg-brand-700 disabled:opacity-50"
              disabled={!product.is_available}
              onClick={handleAdd}
            >
              {product.is_available ? t("product.add") : t("product.unavailable")}
            </button>
          </div>
        </div>
      </div>
    </div>,
    document.body,
  );
}

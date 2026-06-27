import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { Product } from "@/api/types";
import { useCart } from "@/stores/cart";
import { useToast } from "@/stores/toast";
import { formatMoney } from "@/lib/money";
import { ProductModal } from "./ProductModal";

export function ProductCard({ product }: { product: Product }) {
  const { t, i18n } = useTranslation();
  const add = useCart((s) => s.add);
  const toast = useToast((s) => s.show);
  const [showModal, setShowModal] = useState(false);

  const defaultVariant = product.variants?.find((v) => v.is_default) ?? product.variants?.[0];
  const priceFrom = defaultVariant ? defaultVariant.price_cents : product.base_price_cents;

  function onAdd(e: React.MouseEvent) {
    e.stopPropagation();
    if (product.pricing_mode === "weight") {
      add(product, { weightGrams: 500 });
    } else {
      add(product, { variant: defaultVariant });
    }
    toast(t("product.added", "Added to cart!"));
  }

  return (
    <>
      <div
        className="flex flex-col rounded-2xl bg-white shadow-md overflow-hidden cursor-pointer active:scale-[.98] transition-transform"
        onClick={() => setShowModal(true)}
      >
        {/* Circular image on warm gradient */}
        <div className="relative flex justify-center items-center pt-5 pb-3 bg-gradient-to-br from-amber-50 to-brand-100">
          {product.image_url ? (
            <img
              src={product.image_url}
              alt={product.name}
              className="h-24 w-24 rounded-full object-cover shadow-lg ring-4 ring-white"
              loading="lazy"
            />
          ) : (
            <div className="flex h-24 w-24 items-center justify-center rounded-full bg-brand-200 text-4xl shadow-lg ring-4 ring-white">
              🍬
            </div>
          )}
          {product.is_preorder && (
            <span className="absolute top-2 right-2 rounded-full bg-brand-600 px-2 py-0.5 text-xs font-medium text-white">
              {t("product.preorder")}
            </span>
          )}
        </div>

        {/* Text + action */}
        <div className="flex flex-1 flex-col p-3">
          <h3 className="line-clamp-2 min-h-[2.5em] font-bold text-sm leading-snug text-brand-900">
            {product.name}
          </h3>
          {product.description && (
            <p className="line-clamp-2 mt-1 text-xs leading-snug text-brand-600/70">
              {product.description}
            </p>
          )}
          <div className="mt-3 flex items-center justify-between gap-1">
            <span className="font-bold text-sm text-brand-700">
              {formatMoney(priceFrom, "CAD", i18n.language)}
              {product.pricing_mode === "weight" && (
                <span className="text-xs font-normal"> {t("product.perKg")}</span>
              )}
            </span>
            <button
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-brand-600 text-white text-xl leading-none shadow-md transition hover:bg-brand-700 disabled:opacity-40"
              disabled={!product.is_available}
              onClick={onAdd}
              aria-label={t("product.add")}
            >
              +
            </button>
          </div>
          {!product.is_available && (
            <p className="mt-1 text-center text-xs text-brand-400">{t("product.unavailable")}</p>
          )}
        </div>
      </div>

      {showModal && <ProductModal product={product} onClose={() => setShowModal(false)} />}
    </>
  );
}

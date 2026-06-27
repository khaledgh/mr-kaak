import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useCart } from "@/stores/cart";
import { formatMoney } from "@/lib/money";

// Slide-over cart. Quantities are editable; line prices are display estimates
// (the server re-prices at checkout). RTL-aware via logical positioning.
export function CartDrawer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t, i18n } = useTranslation();
  const { lines, setQty, remove, subtotalCents } = useCart();
  const navigate = useNavigate();

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-30" role="dialog" aria-modal="true">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <aside className="absolute inset-y-0 end-0 flex w-full max-w-md flex-col bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-brand-100 p-4">
          <h2 className="text-lg font-bold">{t("cart.title")}</h2>
          <button className="btn-ghost" onClick={onClose} aria-label="Close">✕</button>
        </div>

        <div className="flex-1 space-y-3 overflow-y-auto p-4">
          {lines.length === 0 && <p className="text-brand-700/60">{t("cart.empty")}</p>}
          {lines.map((l) => (
            <div key={l.key} className="card flex items-center gap-3 p-3">
              <div className="flex-1">
                <p className="font-medium">{l.name}</p>
                <p className="text-xs text-brand-700/60">
                  {[l.variantLabel, ...l.modifierLabels].filter(Boolean).join(" · ")}
                  {l.weightGrams ? `${l.weightGrams} g` : ""}
                </p>
                <p className="text-sm font-semibold text-brand-700">
                  {formatMoney(l.unitPriceCents * l.qty, "CAD", i18n.language)}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <button className="btn-ghost px-2" onClick={() => setQty(l.key, l.qty - 1)}>−</button>
                <span className="w-6 text-center">{l.qty}</span>
                <button className="btn-ghost px-2" onClick={() => setQty(l.key, l.qty + 1)}>+</button>
              </div>
              <button className="btn-ghost text-xs text-red-600" onClick={() => remove(l.key)}>
                {t("cart.remove")}
              </button>
            </div>
          ))}
        </div>

        <div className="border-t border-brand-100 p-4">
          <div className="mb-3 flex justify-between font-semibold">
            <span>{t("cart.subtotal")}</span>
            <span>{formatMoney(subtotalCents(), "CAD", i18n.language)}</span>
          </div>
          <button
            className="btn-primary w-full"
            disabled={lines.length === 0}
            onClick={() => { onClose(); navigate("/checkout"); }}
          >
            {t("cart.checkout")}
          </button>
        </div>
      </aside>
    </div>
  );
}

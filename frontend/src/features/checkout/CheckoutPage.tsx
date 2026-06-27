import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { api, qk } from "@/api/endpoints";
import { useCart } from "@/stores/cart";
import { formatMoney } from "@/lib/money";
import type { Address, DeliveryQuote } from "@/api/types";

export function CheckoutPage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { lines, subtotalCents, productIds, clear } = useCart();

  const [fulfillment, setFulfillment] = useState<"delivery" | "pickup">("delivery");
  const [addressId, setAddressId] = useState<number | undefined>();
  const [payment, setPayment] = useState("cod");
  const [coupon, setCoupon] = useState("");
  const [quote, setQuote] = useState<DeliveryQuote | null>(null);
  const [error, setError] = useState("");
  const [placing, setPlacing] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  const addresses = useQuery({ queryKey: qk.addresses(), queryFn: api.addresses });
  const methods = useQuery({ queryKey: ["payment-methods"], queryFn: api.paymentMethods });

  useEffect(() => {
    if (addresses.data?.length && !addressId) setAddressId(addresses.data[0].id);
  }, [addresses.data, addressId]);

  useEffect(() => {
    const addr = addresses.data?.find((a: Address) => a.id === addressId);
    if (fulfillment === "delivery" && addr?.lat != null && addr?.lng != null) {
      api
        .deliveryQuote({ lat: addr.lat, lng: addr.lng, product_ids: productIds() })
        .then(setQuote)
        .catch(() => setQuote(null));
    } else {
      setQuote(null);
    }
  }, [addressId, fulfillment, addresses.data, productIds]);

  const subtotal = subtotalCents();
  const fee = fulfillment === "delivery" ? quote?.fee_cents ?? 0 : 0;
  const total = subtotal + fee;

  // Inline validation: delivery requires an address selection
  const needsAddress = fulfillment === "delivery" && !addressId && (addresses.data ?? []).length === 0;
  const addressError = submitted && needsAddress ? t("validation.selectAddress") : "";

  async function placeOrder() {
    setSubmitted(true);
    if (needsAddress) return;

    setError("");
    setPlacing(true);
    try {
      const body = {
        fulfillment_type: fulfillment,
        address_id: fulfillment === "delivery" ? addressId : undefined,
        payment_method: payment,
        coupon_code: coupon || undefined,
        items: lines.map((l) => ({
          product_id: l.productId,
          variant_id: l.variantId,
          qty: l.qty,
          weight_grams: l.weightGrams,
          modifier_ids: l.modifierIds,
        })),
      };
      const order = await api.checkout(body, crypto.randomUUID());
      clear();
      navigate(`/orders/${order.code}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : t("common.error"));
    } finally {
      setPlacing(false);
    }
  }

  if (lines.length === 0) {
    return <p className="text-brand-700/60">{t("cart.empty")}</p>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("checkout.title")}</h1>

      {/* Fulfillment */}
      <section className="card space-y-3 p-4">
        <h2 className="font-semibold">{t("checkout.fulfillment")}</h2>
        <div className="flex gap-2">
          {(["delivery", "pickup"] as const).map((f) => (
            <button
              key={f}
              className={f === fulfillment ? "btn-primary" : "btn-ghost"}
              onClick={() => { setFulfillment(f); setSubmitted(false); }}
            >
              {t(`checkout.${f}`)}
            </button>
          ))}
        </div>

        {fulfillment === "delivery" && (
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-800">
              {t("checkout.address")}
            </label>
            {(addresses.data ?? []).length > 0 ? (
              <select
                className="input"
                value={addressId}
                onChange={(e) => { setAddressId(Number(e.target.value)); setSubmitted(false); }}
              >
                {(addresses.data ?? []).map((a: Address) => (
                  <option key={a.id} value={a.id}>
                    {a.label ? `${a.label} — ` : ""}{a.line1}, {a.city}
                  </option>
                ))}
              </select>
            ) : (
              <p className="mt-1 text-sm text-red-600">
                {t("checkout.noAddress")}{" "}
                <Link to="/addresses" className="underline">{t("address.add")} →</Link>
              </p>
            )}
            {addressError && <p className="mt-1 text-xs text-red-500">{addressError}</p>}
            {quote && !quote.deliverable && (
              <p className="mt-1 text-sm text-red-600">{quote.reason}</p>
            )}
          </div>
        )}
      </section>

      {/* Payment */}
      <section className="card space-y-2 p-4">
        <h2 className="font-semibold">{t("checkout.payment")}</h2>
        {(methods.data ?? ["cod"]).map((m) => (
          <label key={m} className="flex cursor-pointer items-center gap-2">
            <input type="radio" name="pay" checked={payment === m} onChange={() => setPayment(m)} />
            <span className="text-sm">{m === "cod" ? t("checkout.cod") : t("checkout.card")}</span>
          </label>
        ))}
      </section>

      {/* Coupon + totals */}
      <section className="card space-y-3 p-4">
        <input
          className="input"
          placeholder={t("checkout.coupon")}
          value={coupon}
          onChange={(e) => setCoupon(e.target.value.toUpperCase())}
        />
        <div className="space-y-1">
          <Row label={t("cart.subtotal")} value={formatMoney(subtotal, "CAD", i18n.language)} />
          {fulfillment === "delivery" && (
            <Row label={t("checkout.fee")} value={formatMoney(fee, "CAD", i18n.language)} />
          )}
          <Row label={t("checkout.total")} value={formatMoney(total, "CAD", i18n.language)} bold />
        </div>
      </section>

      {error && <p className="text-red-600">{error}</p>}

      <button
        className="btn-primary w-full"
        disabled={placing}
        onClick={placeOrder}
      >
        {placing ? t("common.loading") : t("checkout.placeOrder")}
      </button>
    </div>
  );
}

function Row({ label, value, bold }: { label: string; value: string; bold?: boolean }) {
  return (
    <div className={`flex justify-between ${bold ? "border-t border-brand-100 pt-2 font-bold" : "text-sm text-brand-700/80"}`}>
      <span>{label}</span>
      <span>{value}</span>
    </div>
  );
}

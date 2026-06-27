import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { api, qk } from "@/api/endpoints";
import { useAuth } from "@/stores/auth";
import { formatMoney } from "@/lib/money";

const STEPS = ["confirmed", "preparing", "ready", "out_for_delivery", "delivered"];

// Live order tracking: loads the order, then subscribes to the SSE stream for
// instant status updates (token passed as query param since EventSource can't
// set headers).
export function OrderTrackPage() {
  const { code = "" } = useParams();
  const { t, i18n } = useTranslation();
  const token = useAuth((s) => s.accessToken);
  const [liveStatus, setLiveStatus] = useState<string | null>(null);

  const order = useQuery({ queryKey: qk.order(code), queryFn: () => api.order(code), enabled: !!code });

  useEffect(() => {
    if (!code || !token) return;
    const es = new EventSource(`/api/v1/orders/${code}/stream?access_token=${encodeURIComponent(token)}`);
    es.addEventListener("status", (e) => {
      try {
        const data = JSON.parse((e as MessageEvent).data) as { status: string };
        setLiveStatus(data.status);
      } catch {
        /* ignore malformed */
      }
    });
    es.onerror = () => es.close(); // fall back to the fetched status
    return () => es.close();
  }, [code, token]);

  if (order.isLoading) return <p>{t("common.loading")}</p>;
  if (order.isError || !order.data) return <p className="text-red-600">{t("common.error")}</p>;

  const o = order.data;
  const status = liveStatus ?? o.status;
  const activeIdx = STEPS.indexOf(status);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">{t("order.tracking")}</h1>
        <p className="text-brand-700/60">#{o.code}</p>
      </div>

      {/* Status stepper */}
      <ol className="space-y-2">
        {STEPS.map((s, i) => (
          <li key={s} className="flex items-center gap-3">
            <span
              className={`flex h-7 w-7 items-center justify-center rounded-full text-sm ${
                i <= activeIdx ? "bg-brand-600 text-white" : "bg-brand-100 text-brand-400"
              }`}
            >
              {i <= activeIdx ? "✓" : i + 1}
            </span>
            <span className={i === activeIdx ? "font-semibold" : ""}>{t(`order.status.${s}`)}</span>
          </li>
        ))}
      </ol>

      {(status === "cancelled" || status === "refunded") && (
        <p className="font-semibold text-red-600">{t(`order.status.${status}`)}</p>
      )}

      <section className="card p-4">
        {o.items.map((it) => (
          <div key={it.id} className="flex justify-between border-b border-brand-50 py-2 last:border-0">
            <span>{it.qty}× {it.name_snapshot}</span>
            <span>{formatMoney(it.line_total_cents, o.currency, i18n.language)}</span>
          </div>
        ))}
        <div className="mt-3 flex justify-between font-bold">
          <span>{t("checkout.total")}</span>
          <span>{formatMoney(o.total_cents, o.currency, i18n.language)}</span>
        </div>
      </section>
    </div>
  );
}

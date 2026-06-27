import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { api, qk } from "@/api/endpoints";
import { formatMoney } from "@/lib/money";

// Customer order history.
export function OrdersPage() {
  const { t, i18n } = useTranslation();
  const orders = useQuery({ queryKey: qk.myOrders(), queryFn: api.myOrders });

  if (orders.isLoading) return <p>{t("common.loading")}</p>;

  return (
    <div className="space-y-3">
      <h1 className="text-2xl font-bold">{t("nav.orders")}</h1>
      {(orders.data ?? []).map((o) => (
        <Link key={o.id} to={`/orders/${o.code}`} className="card flex items-center justify-between p-4">
          <div>
            <p className="font-semibold">#{o.code}</p>
            <p className="text-sm text-brand-700/60">{t(`order.status.${o.status}`)}</p>
          </div>
          <span className="font-bold">{formatMoney(o.total_cents, o.currency, i18n.language)}</span>
        </Link>
      ))}
      {orders.data?.length === 0 && <p className="text-brand-700/60">—</p>}
    </div>
  );
}

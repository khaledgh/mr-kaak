import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import type { Order, OrderStatus } from "@/api/types";
import { STATUS_COLORS, STATUS_LABELS } from "@/api/types";

const FILTER_STATUSES: Array<{ value: string; label: string }> = [
  { value: "", label: "All" },
  { value: "pending_payment", label: "Pending" },
  { value: "confirmed", label: "Confirmed" },
  { value: "preparing", label: "Preparing" },
  { value: "ready", label: "Ready" },
  { value: "out_for_delivery", label: "Delivering" },
  { value: "delivered", label: "Delivered" },
  { value: "cancelled", label: "Cancelled" },
];

function fmt(cents: number) {
  return new Intl.NumberFormat("en-CA", { style: "currency", currency: "CAD" }).format(cents / 100);
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

export function OrdersPage() {
  const [status, setStatus] = useState("");

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.adminOrders(status),
    queryFn: () => api.adminOrders({ status: status || undefined, per_page: 50 }),
    refetchInterval: 30_000,
  });

  const orders = data?.orders ?? [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold">Orders</h1>
        <button className="btn-ghost text-sm" onClick={() => refetch()}>Refresh</button>
      </div>

      {/* Status filter tabs */}
      <div className="flex flex-wrap gap-2">
        {FILTER_STATUSES.map((f) => (
          <button
            key={f.value}
            onClick={() => setStatus(f.value)}
            className={`rounded-full px-3 py-1 text-sm font-medium transition ${
              status === f.value
                ? "bg-brand-600 text-white"
                : "bg-white text-gray-600 ring-1 ring-gray-200 hover:bg-gray-50"
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {isLoading && <p className="text-gray-500">Loading orders…</p>}
      {error && (
        <p className="rounded bg-red-50 px-4 py-2 text-sm text-red-600">
          {error instanceof Error ? error.message : "Failed to load orders."}
        </p>
      )}

      {!isLoading && orders.length === 0 && (
        <p className="py-12 text-center text-gray-400">No orders found.</p>
      )}

      <div className="space-y-2">
        {orders.map((order: Order) => (
          <Link
            key={order.id}
            to={`/orders/${order.code}`}
            className="card flex items-center gap-4 p-4 hover:bg-gray-50 transition-colors"
          >
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2">
                <span className="font-mono font-semibold text-brand-700">{order.code}</span>
                <StatusBadge status={order.status} />
                <span className={`rounded px-2 py-0.5 text-xs ${
                  order.fulfillment_type === "delivery"
                    ? "bg-blue-50 text-blue-600"
                    : "bg-gray-100 text-gray-600"
                }`}>
                  {order.fulfillment_type}
                </span>
              </div>
              {order.address_snapshot && (
                <p className="mt-0.5 truncate text-sm text-gray-500">
                  {order.address_snapshot.line1}, {order.address_snapshot.city}
                </p>
              )}
            </div>
            <div className="shrink-0 text-right">
              <p className="font-semibold">{fmt(order.total_cents)}</p>
              <p className="text-xs text-gray-400">{timeAgo(order.created_at)}</p>
            </div>
            <svg className="h-4 w-4 shrink-0 text-gray-300" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        ))}
      </div>

      {data?.meta && data.meta.total > orders.length && (
        <p className="text-center text-sm text-gray-400">
          Showing {orders.length} of {data.meta.total} orders
        </p>
      )}
    </div>
  );
}

export function StatusBadge({ status }: { status: OrderStatus }) {
  return (
    <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[status] ?? "bg-gray-100 text-gray-600"}`}>
      {STATUS_LABELS[status] ?? status}
    </span>
  );
}

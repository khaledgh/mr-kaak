import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import type { OrderStatus } from "@/api/types";
import { STATUS_TRANSITIONS, STATUS_LABELS } from "@/api/types";
import { StatusBadge } from "./OrdersPage";

function fmt(cents: number) {
  return new Intl.NumberFormat("en-CA", { style: "currency", currency: "CAD" }).format(cents / 100);
}

function fmtDate(dateStr: string) {
  return new Date(dateStr).toLocaleString("en-CA", {
    year: "numeric", month: "short", day: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

export function OrderDetailPage() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [note, setNote] = useState("");
  const [actionError, setActionError] = useState("");

  const { data: order, isLoading, error } = useQuery({
    queryKey: qk.adminOrder(code!),
    queryFn: () => api.adminOrder(code!),
    enabled: !!code,
    refetchInterval: 15_000,
  });

  const advanceMutation = useMutation({
    mutationFn: ({ status }: { status: string }) =>
      api.advanceStatus(order!.id, status, note || undefined),
    onSuccess: (updated) => {
      qc.setQueryData(qk.adminOrder(code!), updated);
      qc.invalidateQueries({ queryKey: ["admin-orders"] });
      setNote("");
      setActionError("");
    },
    onError: (e) => setActionError(e instanceof Error ? e.message : "Failed to advance status."),
  });

  if (isLoading) {
    return <p className="text-gray-500">Loading order…</p>;
  }

  if (error || !order) {
    return (
      <div className="space-y-4">
        <Link to="/orders" className="text-sm text-brand-600 hover:underline">← Orders</Link>
        <p className="text-red-600">{error instanceof Error ? error.message : "Order not found."}</p>
      </div>
    );
  }

  const nextStatuses = STATUS_TRANSITIONS[order.status] ?? [];

  return (
    <div className="space-y-6">
      {/* Back + header */}
      <div className="flex flex-wrap items-start gap-3">
        <button className="btn-ghost text-sm" onClick={() => navigate("/orders")}>
          ← Orders
        </button>
        <div className="flex-1 min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="font-mono text-xl font-bold text-brand-700">{order.code}</h1>
            <StatusBadge status={order.status} />
            <span className={`rounded px-2 py-0.5 text-xs ${
              order.fulfillment_type === "delivery"
                ? "bg-blue-50 text-blue-600"
                : "bg-gray-100 text-gray-600"
            }`}>
              {order.fulfillment_type}
            </span>
          </div>
          <p className="mt-0.5 text-sm text-gray-500">
            Placed {fmtDate(order.created_at)} · Customer #{order.user_id}
          </p>
        </div>
      </div>

      {/* Advance status */}
      {nextStatuses.length > 0 && (
        <section className="card space-y-3 p-4">
          <h2 className="font-semibold">Advance Status</h2>
          <textarea
            className="input resize-none text-sm"
            rows={2}
            placeholder="Optional note (e.g. driver assigned, reason for cancellation)…"
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
          {actionError && (
            <p className="rounded bg-red-50 px-3 py-2 text-sm text-red-600">{actionError}</p>
          )}
          <div className="flex flex-wrap gap-2">
            {nextStatuses.map((s: OrderStatus) => (
              <button
                key={s}
                className={s === "cancelled" || s === "refunded" ? "btn-danger" : "btn-primary"}
                disabled={advanceMutation.isPending}
                onClick={() => advanceMutation.mutate({ status: s })}
              >
                → {STATUS_LABELS[s]}
              </button>
            ))}
          </div>
        </section>
      )}

      {/* Address snapshot */}
      {order.address_snapshot && (
        <section className="card p-4">
          <h2 className="mb-2 font-semibold">Delivery Address</h2>
          <div className="text-sm text-gray-700 space-y-0.5">
            {order.address_snapshot.label && (
              <p className="font-medium">{order.address_snapshot.label}</p>
            )}
            <p>{order.address_snapshot.line1}{order.address_snapshot.line2 ? `, ${order.address_snapshot.line2}` : ""}</p>
            <p>{order.address_snapshot.city}, {order.address_snapshot.province_code} {order.address_snapshot.postal_code}</p>
            {order.address_snapshot.phone && <p>{order.address_snapshot.phone}</p>}
            {order.address_snapshot.notes && (
              <p className="text-gray-500 italic">{order.address_snapshot.notes}</p>
            )}
          </div>
        </section>
      )}

      {/* Items */}
      {order.items && order.items.length > 0 && (
        <section className="card overflow-hidden">
          <h2 className="border-b border-gray-100 px-4 py-3 font-semibold">Items</h2>
          <table className="w-full text-sm">
            <thead className="bg-gray-50 text-left text-xs uppercase text-gray-500">
              <tr>
                <th className="px-4 py-2">Item</th>
                <th className="px-4 py-2 text-center">Qty</th>
                <th className="px-4 py-2 text-right">Unit</th>
                <th className="px-4 py-2 text-right">Total</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {order.items.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <p className="font-medium">{item.name_snapshot}</p>
                    {item.weight_grams && (
                      <p className="text-xs text-gray-400">{item.weight_grams}g</p>
                    )}
                    {item.modifiers?.map((m) => (
                      <p key={m.modifier_id} className="text-xs text-gray-400">
                        + {m.label}
                        {m.price_delta_cents !== 0 && ` (${fmt(m.price_delta_cents)})`}
                      </p>
                    ))}
                  </td>
                  <td className="px-4 py-3 text-center">{item.qty}</td>
                  <td className="px-4 py-3 text-right">{fmt(item.unit_price_cents)}</td>
                  <td className="px-4 py-3 text-right font-medium">{fmt(item.line_total_cents)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}

      {/* Price summary */}
      <section className="card p-4 space-y-2 text-sm">
        <h2 className="font-semibold">Summary</h2>
        <Row label="Subtotal" value={fmt(order.subtotal_cents)} />
        {order.delivery_fee_cents > 0 && (
          <Row label="Delivery fee" value={fmt(order.delivery_fee_cents)} />
        )}
        {order.discount_cents > 0 && (
          <Row label="Discount" value={`-${fmt(order.discount_cents)}`} />
        )}
        {order.tax_cents > 0 && <Row label="Tax" value={fmt(order.tax_cents)} />}
        <Row label="Total" value={fmt(order.total_cents)} bold />
        <div className="border-t border-gray-100 pt-2 text-xs text-gray-400">
          Payment: {order.payment_method} · Status: {order.payment_status}
        </div>
        {order.notes && (
          <p className="rounded bg-amber-50 px-3 py-2 text-xs text-amber-700">
            Note: {order.notes}
          </p>
        )}
      </section>

      {/* Status history */}
      {order.history && order.history.length > 0 && (
        <section className="card p-4">
          <h2 className="mb-3 font-semibold">History</h2>
          <ol className="space-y-2 text-sm">
            {[...order.history].reverse().map((h) => (
              <li key={h.id} className="flex items-start gap-3">
                <div className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-brand-400" />
                <div>
                  <span className="font-medium">
                    {h.from_status
                      ? `${STATUS_LABELS[h.from_status as OrderStatus] ?? h.from_status} → `
                      : ""}
                    {STATUS_LABELS[h.to_status as OrderStatus] ?? h.to_status}
                  </span>
                  {h.note && <p className="text-gray-500 italic">{h.note}</p>}
                  <p className="text-xs text-gray-400">{fmtDate(h.created_at)}</p>
                </div>
              </li>
            ))}
          </ol>
        </section>
      )}
    </div>
  );
}

function Row({ label, value, bold }: { label: string; value: string; bold?: boolean }) {
  return (
    <div className={`flex justify-between ${bold ? "border-t border-gray-100 pt-2 font-semibold" : ""}`}>
      <span className={bold ? "" : "text-gray-500"}>{label}</span>
      <span>{value}</span>
    </div>
  );
}

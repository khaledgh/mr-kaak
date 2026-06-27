import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import type { DeliveryZone } from "@/api/types";

function c2d(c?: number) { return c != null ? (c / 100).toFixed(2) : "0"; }

export function DeliveryPage() {
  const navigate = useNavigate();
  const [toDelete, setToDelete] = useState<DeliveryZone | null>(null);
  const { show } = useToast();
  const qc = useQueryClient();

  const { data: zones = [], isLoading } = useQuery({ queryKey: qk.zones(), queryFn: api.deliveryZones });

  const del = useMutation({
    mutationFn: (id: number) => api.deleteZone(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.zones() }); setToDelete(null); show("Deleted"); },
    onError: () => show("Delete failed", "error"),
  });

  return (
    <div>
      <PageHeader title="Delivery Zones" action={<button className="btn-primary" onClick={() => navigate("/delivery/new")}>+ New zone</button>} />

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-12 rounded-xl" />)}</div>
      ) : !zones.length ? (
        <EmptyState icon="🚚" title="No delivery zones" description="Define areas where delivery is available." action={<button className="btn-primary" onClick={() => navigate("/delivery/new")}>Add zone</button>} />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead><tr><th>Zone</th><th>Shape</th><th>Fee</th><th>Min order</th><th>Active</th><th></th></tr></thead>
            <tbody>
              {zones.map((z) => (
                <tr key={z.id}>
                  <td className="font-medium">{z.name}</td>
                  <td className="capitalize text-xs text-gray-500">{z.shape}{z.shape === "radius" ? ` · ${z.radius_km} km` : ""}</td>
                  <td>${c2d(z.fee_cents)}</td>
                  <td>${c2d(z.min_order_cents)}</td>
                  <td><span className={`badge ${z.is_active ? "badge-green" : "badge-gray"}`}>{z.is_active ? "Active" : "Off"}</span></td>
                  <td><div className="flex gap-1">
                    <button className="btn-ghost py-1 px-2 text-xs" onClick={() => navigate(`/delivery/${z.id}/edit`)}>Edit</button>
                    <button className="btn-ghost py-1 px-2 text-xs text-red-500" onClick={() => setToDelete(z)}>Delete</button>
                  </div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <ConfirmDialog
        open={!!toDelete} title="Delete zone?" message={`Delete "${toDelete?.name}"?`}
        confirmLabel="Delete" danger loading={del.isPending}
        onConfirm={() => toDelete && del.mutate(toDelete.id)}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}

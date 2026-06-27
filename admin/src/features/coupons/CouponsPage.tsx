import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { Modal } from "@/components/Modal";
import { FormField } from "@/components/FormField";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import type { Coupon } from "@/api/types";

function blankForm(): Partial<Coupon> & { value_str: string; min_str: string; max_str: string } {
  return {
    code: "", type: "percent", value: 0, value_str: "", min_order_cents: 0, min_str: "0",
    max_discount_cents: undefined, max_str: "", usage_limit: undefined, per_user_limit: undefined,
    starts_at: "", ends_at: "", is_active: true,
  };
}

function c2d(cents?: number) { return cents != null ? (cents / 100).toFixed(2) : ""; }
function d2c(s: string) { return s ? Math.round(parseFloat(s) * 100) : 0; }
function toISO(local?: string) {
  if (!local) return undefined;
  const d = new Date(local);
  return isNaN(d.getTime()) ? undefined : d.toISOString();
}
function toLocal(iso?: string) {
  if (!iso) return "";
  const d = new Date(iso);
  return isNaN(d.getTime()) ? "" : d.toISOString().slice(0, 16);
}

export function CouponsPage() {
  const [form, setForm] = useState(blankForm());
  const [editing, setEditing] = useState<Coupon | null>(null);
  const [open, setOpen] = useState(false);
  const [toDelete, setToDelete] = useState<Coupon | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { show } = useToast();
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({ queryKey: qk.coupons(), queryFn: () => api.coupons({ per_page: 100 }) });

  function openNew() { setForm(blankForm()); setEditing(null); setErrors({}); setOpen(true); }
  function openEdit(c: Coupon) {
    setForm({
      ...c,
      value_str: c.type === "fixed" ? c2d(c.value) : String(c.value),
      min_str: c2d(c.min_order_cents),
      max_str: c2d(c.max_discount_cents),
      starts_at: toLocal(c.starts_at),
      ends_at: toLocal(c.ends_at),
    });
    setEditing(c);
    setErrors({});
    setOpen(true);
  }

  function validate() {
    const e: Record<string, string> = {};
    if (!form.code?.trim()) e.code = "Code is required";
    if (!form.value_str) e.value = "Value is required";
    setErrors(e);
    return !Object.keys(e).length;
  }

  const save = useMutation({
    mutationFn: (body: Partial<Coupon>) => editing ? api.updateCoupon(editing.id, body) : api.createCoupon(body),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.coupons() }); setOpen(false); show(editing ? "Updated" : "Created"); },
    onError: (e) => show(e instanceof Error ? e.message : "Save failed", "error"),
  });

  const del = useMutation({
    mutationFn: (id: number) => api.deleteCoupon(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.coupons() }); setToDelete(null); show("Deleted"); },
    onError: () => show("Delete failed", "error"),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;
    save.mutate({
      code: form.code?.toUpperCase(),
      type: form.type,
      value: form.type === "fixed" ? d2c(form.value_str ?? "0") : Math.round(parseFloat(form.value_str ?? "0")),
      min_order_cents: d2c(form.min_str ?? "0"),
      max_discount_cents: form.max_str ? d2c(form.max_str) : undefined,
      usage_limit: form.usage_limit || undefined,
      per_user_limit: form.per_user_limit || undefined,
      starts_at: toISO(form.starts_at),
      ends_at: toISO(form.ends_at),
      is_active: form.is_active,
    });
  }

  return (
    <div>
      <PageHeader title="Coupons" action={<button className="btn-primary" onClick={openNew}>+ New coupon</button>} />

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 4 }).map((_, i) => <div key={i} className="skeleton h-12 rounded-xl" />)}</div>
      ) : !data?.items.length ? (
        <EmptyState icon="🏷️" title="No coupons yet" action={<button className="btn-primary" onClick={openNew}>Create one</button>} />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead><tr><th>Code</th><th>Type</th><th>Value</th><th>Uses</th><th>Active</th><th>Expires</th><th></th></tr></thead>
            <tbody>
              {data.items.map((c) => (
                <tr key={c.id}>
                  <td className="font-mono font-semibold">{c.code}</td>
                  <td className="capitalize text-xs">{c.type}</td>
                  <td>{c.type === "percent" ? `${c.value}%` : c.type === "fixed" ? `$${c2d(c.value)}` : "Free delivery"}</td>
                  <td className="text-xs text-gray-500">{c.used_count}{c.usage_limit ? `/${c.usage_limit}` : ""}</td>
                  <td><span className={`badge ${c.is_active ? "badge-green" : "badge-gray"}`}>{c.is_active ? "Active" : "Off"}</span></td>
                  <td className="text-xs text-gray-400">{c.ends_at ? new Date(c.ends_at).toLocaleDateString() : "—"}</td>
                  <td><div className="flex gap-1">
                    <button className="btn-ghost py-1 px-2 text-xs" onClick={() => openEdit(c)}>Edit</button>
                    <button className="btn-ghost py-1 px-2 text-xs text-red-500" onClick={() => setToDelete(c)}>Delete</button>
                  </div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal open={open} onClose={() => setOpen(false)} title={editing ? "Edit Coupon" : "New Coupon"} size="md">
        <form onSubmit={submit} className="space-y-4" noValidate>
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Code" required error={errors.code}>
              <input className="input uppercase" value={form.code} onChange={(e) => setForm((f) => ({ ...f, code: e.target.value.toUpperCase() }))} />
            </FormField>
            <FormField label="Type">
              <select className="input" value={form.type} onChange={(e) => setForm((f) => ({ ...f, type: e.target.value as Coupon["type"] }))}>
                <option value="percent">Percent off</option>
                <option value="fixed">Fixed amount off</option>
                <option value="free_delivery">Free delivery</option>
              </select>
            </FormField>
            {form.type !== "free_delivery" && (
              <FormField label={form.type === "percent" ? "Discount %" : "Discount ($)"} required error={errors.value}>
                <input type="number" min={0} step={form.type === "percent" ? 1 : 0.01} className="input" value={form.value_str} onChange={(e) => setForm((f) => ({ ...f, value_str: e.target.value }))} />
              </FormField>
            )}
            <FormField label="Min order ($)">
              <input type="number" min={0} step={0.01} className="input" value={form.min_str} onChange={(e) => setForm((f) => ({ ...f, min_str: e.target.value }))} />
            </FormField>
            {form.type === "percent" && (
              <FormField label="Max discount ($)" hint="Leave blank for no cap">
                <input type="number" min={0} step={0.01} className="input" value={form.max_str} onChange={(e) => setForm((f) => ({ ...f, max_str: e.target.value }))} />
              </FormField>
            )}
            <FormField label="Total usage limit" hint="Blank = unlimited">
              <input type="number" min={0} className="input" value={form.usage_limit ?? ""} onChange={(e) => setForm((f) => ({ ...f, usage_limit: e.target.value ? Number(e.target.value) : undefined }))} />
            </FormField>
            <FormField label="Per-user limit">
              <input type="number" min={0} className="input" value={form.per_user_limit ?? ""} onChange={(e) => setForm((f) => ({ ...f, per_user_limit: e.target.value ? Number(e.target.value) : undefined }))} />
            </FormField>
            <FormField label="Starts at">
              <input type="datetime-local" className="input" value={form.starts_at ?? ""} onChange={(e) => setForm((f) => ({ ...f, starts_at: e.target.value }))} />
            </FormField>
            <FormField label="Ends at">
              <input type="datetime-local" className="input" value={form.ends_at ?? ""} onChange={(e) => setForm((f) => ({ ...f, ends_at: e.target.value }))} />
            </FormField>
          </div>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <button type="button" className={`switch ${form.is_active ? "bg-brand-500" : "bg-gray-200"}`} onClick={() => setForm((f) => ({ ...f, is_active: !f.is_active }))}>
              <span className={`switch-thumb ${form.is_active ? "translate-x-4" : "translate-x-0"}`} />
            </button>
            Active
          </label>
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" className="btn-ghost" onClick={() => setOpen(false)}>Cancel</button>
            <button type="submit" className="btn-primary" disabled={save.isPending}>{save.isPending ? "Saving…" : "Save"}</button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!toDelete} title="Delete coupon?" message={`Delete "${toDelete?.code}"?`}
        confirmLabel="Delete" danger loading={del.isPending}
        onConfirm={() => toDelete && del.mutate(toDelete.id)}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}

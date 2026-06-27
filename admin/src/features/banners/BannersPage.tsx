import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { Modal } from "@/components/Modal";
import { FormField } from "@/components/FormField";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { MediaPicker } from "@/components/MediaPicker";
import type { Banner } from "@/api/types";

function blankForm(): Partial<Banner> {
  return { title: "", image_url: "", link_url: "", sort_order: 0, starts_at: "", ends_at: "", is_active: true };
}

export function BannersPage() {
  const [form, setForm] = useState(blankForm());
  const [editing, setEditing] = useState<Banner | null>(null);
  const [open, setOpen] = useState(false);
  const [toDelete, setToDelete] = useState<Banner | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { show } = useToast();
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({ queryKey: qk.banners(), queryFn: () => api.banners({ per_page: 100 }) });

  function openNew() { setForm(blankForm()); setEditing(null); setErrors({}); setOpen(true); }
  function openEdit(b: Banner) { setForm({ ...b }); setEditing(b); setErrors({}); setOpen(true); }

  function validate() {
    const e: Record<string, string> = {};
    if (!form.image_url?.trim()) e.image_url = "Image is required";
    setErrors(e);
    return !Object.keys(e).length;
  }

  const save = useMutation({
    mutationFn: (body: Partial<Banner>) => editing ? api.updateBanner(editing.id, body) : api.createBanner(body),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.banners() }); setOpen(false); show(editing ? "Updated" : "Created"); },
    onError: (e) => show(e instanceof Error ? e.message : "Save failed", "error"),
  });

  const del = useMutation({
    mutationFn: (id: number) => api.deleteBanner(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.banners() }); setToDelete(null); show("Deleted"); },
    onError: () => show("Delete failed", "error"),
  });

  function toISO(local?: string) {
    if (!local) return undefined;
    const d = new Date(local);
    return isNaN(d.getTime()) ? undefined : d.toISOString();
  }

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;
    save.mutate({
      title: form.title || undefined,
      image_url: form.image_url!,
      link_url: form.link_url || undefined,
      sort_order: form.sort_order ?? 0,
      starts_at: toISO(form.starts_at),
      ends_at: toISO(form.ends_at),
      is_active: form.is_active,
    });
  }

  return (
    <div>
      <PageHeader title="Banners" action={<button className="btn-primary" onClick={openNew}>+ New banner</button>} />

      {isLoading ? (
        <div className="grid grid-cols-2 gap-3">{Array.from({ length: 4 }).map((_, i) => <div key={i} className="skeleton aspect-video rounded-xl" />)}</div>
      ) : !data?.items.length ? (
        <EmptyState icon="📢" title="No banners yet" action={<button className="btn-primary" onClick={openNew}>Add banner</button>} />
      ) : (
        <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
          {data.items.map((b) => (
            <div key={b.id} className="card overflow-hidden">
              <div className="relative aspect-video bg-gray-100">
                <img src={b.image_url} alt={b.title ?? ""} className="h-full w-full object-cover" />
                {!b.is_active && (
                  <span className="absolute right-2 top-2 rounded bg-black/60 px-1.5 py-0.5 text-xs text-white">Draft</span>
                )}
              </div>
              <div className="p-3">
                {b.title && <p className="text-sm font-medium">{b.title}</p>}
                {b.link_url && <p className="truncate text-xs text-gray-400">{b.link_url}</p>}
                <div className="mt-2 flex gap-1">
                  <button className="btn-ghost py-1 px-2 text-xs" onClick={() => openEdit(b)}>Edit</button>
                  <button className="btn-ghost py-1 px-2 text-xs text-red-500" onClick={() => setToDelete(b)}>Delete</button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal open={open} onClose={() => setOpen(false)} title={editing ? "Edit Banner" : "New Banner"} size="md">
        <form onSubmit={submit} className="space-y-4" noValidate>
          <MediaPicker value={form.image_url} onChange={(url) => setForm((f) => ({ ...f, image_url: url }))} label="Banner image *" />
          {errors.image_url && <p className="field-error">{errors.image_url}</p>}
          <FormField label="Title"><input className="input" value={form.title ?? ""} onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))} /></FormField>
          <FormField label="Link URL"><input type="url" className="input" placeholder="https://…" value={form.link_url ?? ""} onChange={(e) => setForm((f) => ({ ...f, link_url: e.target.value }))} /></FormField>
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Sort order"><input type="number" className="input" value={form.sort_order ?? 0} onChange={(e) => setForm((f) => ({ ...f, sort_order: Number(e.target.value) }))} /></FormField>
            <div className="flex items-end pb-0.5">
              <label className="flex cursor-pointer items-center gap-2 text-sm">
                <button type="button" className={`switch ${form.is_active ? "bg-brand-500" : "bg-gray-200"}`} onClick={() => setForm((f) => ({ ...f, is_active: !f.is_active }))}>
                  <span className={`switch-thumb ${form.is_active ? "translate-x-4" : "translate-x-0"}`} />
                </button>
                Active
              </label>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Starts at"><input type="datetime-local" className="input" value={form.starts_at ?? ""} onChange={(e) => setForm((f) => ({ ...f, starts_at: e.target.value }))} /></FormField>
            <FormField label="Ends at"><input type="datetime-local" className="input" value={form.ends_at ?? ""} onChange={(e) => setForm((f) => ({ ...f, ends_at: e.target.value }))} /></FormField>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" className="btn-ghost" onClick={() => setOpen(false)}>Cancel</button>
            <button type="submit" className="btn-primary" disabled={save.isPending}>{save.isPending ? "Saving…" : "Save"}</button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!toDelete} title="Delete banner?" message={`Delete "${toDelete?.title || "this banner"}"?`}
        confirmLabel="Delete" danger loading={del.isPending}
        onConfirm={() => toDelete && del.mutate(toDelete.id)}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}

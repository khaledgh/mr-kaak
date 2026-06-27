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
import type { Category } from "@/api/types";

const LOCALES = ["en", "ar", "fr"] as const;

type LocaleKey = (typeof LOCALES)[number];

const BLANK = {
  slug: "", image_url: "", sort_order: 0, is_active: true,
  translations: { en: { name: "", description: "" }, ar: { name: "", description: "" }, fr: { name: "", description: "" } },
};

export function CategoriesPage() {
  const [form, setForm] = useState({ ...BLANK });
  const [editing, setEditing] = useState<Category | null>(null);
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<LocaleKey>("en");
  const [toDelete, setToDelete] = useState<Category | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { show } = useToast();
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: qk.categories(),
    queryFn: () => api.categories({ per_page: 100 }),
  });

  function openNew() {
    setForm({ ...BLANK });
    setEditing(null);
    setErrors({});
    setTab("en");
    setOpen(true);
  }

  function openEdit(cat: Category) {
    const tr = { en: { name: "", description: "" }, ar: { name: "", description: "" }, fr: { name: "", description: "" } };
    for (const l of LOCALES) {
      if (cat.translations?.[l]) tr[l] = { name: cat.translations[l].name ?? "", description: cat.translations[l].description ?? "" };
    }
    setForm({ slug: cat.slug, image_url: cat.image_url ?? "", sort_order: cat.sort_order, is_active: cat.is_active, translations: tr });
    setEditing(cat);
    setErrors({});
    setTab("en");
    setOpen(true);
  }

  function validate() {
    const e: Record<string, string> = {};
    if (!form.slug.trim()) e.slug = "Slug is required";
    if (!form.translations.en.name.trim()) e.name_en = "English name is required";
    setErrors(e);
    return Object.keys(e).length === 0;
  }

  const save = useMutation({
    mutationFn: (body: Partial<Category>) =>
      editing ? api.updateCategory(editing.id, body) : api.createCategory(body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.categories() });
      setOpen(false);
      show(editing ? "Category updated" : "Category created");
    },
    onError: (e) => show(e instanceof Error ? e.message : "Save failed", "error"),
  });

  const del = useMutation({
    mutationFn: (id: number) => api.deleteCategory(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.categories() }); setToDelete(null); show("Deleted"); },
    onError: (e) => show(e instanceof Error ? e.message : "Delete failed", "error"),
  });

  const toggle = useMutation({
    mutationFn: ({ id, v }: { id: number; v: boolean }) => api.toggleCategory(id, v),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.categories() }),
    onError: () => show("Toggle failed", "error"),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;
    const tr: Category["translations"] = {};
    for (const l of LOCALES) {
      if (form.translations[l].name.trim()) tr[l] = { name: form.translations[l].name, description: form.translations[l].description || undefined };
    }
    save.mutate({ slug: form.slug, image_url: form.image_url || undefined, sort_order: form.sort_order, is_active: form.is_active, translations: tr });
  }

  return (
    <div>
      <PageHeader
        title="Categories"
        action={<button className="btn-primary" onClick={openNew}>+ New category</button>}
      />

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 4 }).map((_, i) => <div key={i} className="skeleton h-14 rounded-xl" />)}</div>
      ) : !data?.items.length ? (
        <EmptyState icon="🍬" title="No categories yet" action={<button className="btn-primary" onClick={openNew}>Create one</button>} />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Category</th>
                <th>Slug</th>
                <th>Order</th>
                <th>Active</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((cat) => (
                <tr key={cat.id}>
                  <td>
                    <div className="flex items-center gap-3">
                      {cat.image_url && <img src={cat.image_url} alt="" className="h-9 w-9 rounded-lg object-cover" />}
                      <span className="font-medium">{cat.translations?.en?.name ?? cat.slug}</span>
                    </div>
                  </td>
                  <td className="font-mono text-xs text-gray-400">{cat.slug}</td>
                  <td>{cat.sort_order}</td>
                  <td>
                    <button
                      className={`switch ${cat.is_active ? "bg-brand-500" : "bg-gray-200"}`}
                      onClick={() => toggle.mutate({ id: cat.id, v: !cat.is_active })}
                    >
                      <span className={`switch-thumb ${cat.is_active ? "translate-x-4" : "translate-x-0"}`} />
                    </button>
                  </td>
                  <td>
                    <div className="flex gap-1">
                      <button className="btn-ghost py-1 px-2 text-xs" onClick={() => openEdit(cat)}>Edit</button>
                      <button className="btn-ghost py-1 px-2 text-xs text-red-500" onClick={() => setToDelete(cat)}>Delete</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal open={open} onClose={() => setOpen(false)} title={editing ? "Edit Category" : "New Category"} size="lg">
        <form onSubmit={submit} className="space-y-4" noValidate>
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Slug" required error={errors.slug}>
              <input className={`input ${errors.slug ? "border-red-400" : ""}`} value={form.slug} onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value }))} placeholder="arabic-sweets" />
            </FormField>
            <FormField label="Sort order">
              <input type="number" className="input" value={form.sort_order} onChange={(e) => setForm((f) => ({ ...f, sort_order: Number(e.target.value) }))} />
            </FormField>
          </div>

          <div className="flex items-center gap-3">
            <button type="button" className={`switch ${form.is_active ? "bg-brand-500" : "bg-gray-200"}`} onClick={() => setForm((f) => ({ ...f, is_active: !f.is_active }))}>
              <span className={`switch-thumb ${form.is_active ? "translate-x-4" : "translate-x-0"}`} />
            </button>
            <span className="text-sm text-gray-600">Active</span>
          </div>

          <MediaPicker value={form.image_url} onChange={(url) => setForm((f) => ({ ...f, image_url: url }))} />

          {/* Translations tabs */}
          <div>
            <div className="mb-2 flex gap-1 border-b border-gray-100">
              {LOCALES.map((l) => (
                <button key={l} type="button" className={`px-3 py-1.5 text-sm font-medium transition ${tab === l ? "border-b-2 border-brand-500 text-brand-600" : "text-gray-400 hover:text-gray-700"}`} onClick={() => setTab(l)}>
                  {l.toUpperCase()}
                </button>
              ))}
            </div>
            <div className="space-y-3">
              <FormField label="Name" required={tab === "en"} error={tab === "en" ? errors.name_en : undefined}>
                <input className={`input ${tab === "en" && errors.name_en ? "border-red-400" : ""}`} value={form.translations[tab].name} onChange={(e) => setForm((f) => ({ ...f, translations: { ...f.translations, [tab]: { ...f.translations[tab], name: e.target.value } } }))} />
              </FormField>
              <FormField label="Description">
                <input className="input" value={form.translations[tab].description ?? ""} onChange={(e) => setForm((f) => ({ ...f, translations: { ...f.translations, [tab]: { ...f.translations[tab], description: e.target.value } } }))} />
              </FormField>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" className="btn-ghost" onClick={() => setOpen(false)}>Cancel</button>
            <button type="submit" className="btn-primary" disabled={save.isPending}>{save.isPending ? "Saving…" : "Save"}</button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!toDelete}
        title="Delete category?"
        message={`Delete "${toDelete?.translations?.en?.name ?? toDelete?.slug}"? This will fail if the category has products.`}
        confirmLabel="Delete"
        danger
        loading={del.isPending}
        onConfirm={() => toDelete && del.mutate(toDelete.id)}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}

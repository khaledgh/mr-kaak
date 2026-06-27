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
import type { ModifierGroup, Product, ProductVariant } from "@/api/types";

const LOCALES = ["en", "ar", "fr"] as const;
type LocaleKey = (typeof LOCALES)[number];

const BLANK_VARIANT: ProductVariant = { label: "", sku: "", price_cents: 0, is_default: false, sort_order: 0 };
const BLANK_MOD: import("@/api/types").Modifier = { label: "", price_delta_cents: 0, is_default: false, is_available: true };
const BLANK_GROUP: ModifierGroup = { label: "", min_select: 0, max_select: 1, is_required: false, sort_order: 0, modifiers: [{ ...BLANK_MOD }] };

function blankForm() {
  return {
    category_id: 0,
    slug: "",
    image_url: "",
    pricing_mode: "unit" as Product["pricing_mode"],
    base_price_cents: 0,
    is_preorder: false,
    preorder_lead_hours: 24,
    is_available: true,
    allergens: [] as string[],
    sort_order: 0,
    translations: { en: { name: "", description: "" }, ar: { name: "", description: "" }, fr: { name: "", description: "" } },
    variants: [] as ProductVariant[],
    modifier_groups: [] as ModifierGroup[],
  };
}

export function ProductsPage() {
  const [selectedCat, setSelectedCat] = useState(0);
  const [form, setForm] = useState(blankForm());
  const [editing, setEditing] = useState<Product | null>(null);
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<LocaleKey>("en");
  const [toDelete, setToDelete] = useState<Product | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { show } = useToast();
  const qc = useQueryClient();

  const cats = useQuery({ queryKey: qk.categories(), queryFn: () => api.categories({ per_page: 100 }) });
  const { data, isLoading } = useQuery({
    queryKey: qk.products(selectedCat || undefined),
    queryFn: () => api.products({ category_id: selectedCat || undefined, per_page: 100 }),
  });

  function openNew() {
    setForm(blankForm());
    setEditing(null);
    setErrors({});
    setTab("en");
    setOpen(true);
  }

  function openEdit(p: Product) {
    const tr = { en: { name: "", description: "" }, ar: { name: "", description: "" }, fr: { name: "", description: "" } };
    for (const l of LOCALES) {
      if (p.translations?.[l]) tr[l] = { name: p.translations[l].name ?? "", description: p.translations[l].description ?? "" };
    }
    setForm({
      category_id: p.category_id,
      slug: p.slug,
      image_url: p.image_url ?? "",
      pricing_mode: p.pricing_mode,
      base_price_cents: p.base_price_cents,
      is_preorder: p.is_preorder,
      preorder_lead_hours: p.preorder_lead_hours ?? 24,
      is_available: p.is_available,
      allergens: p.allergens ?? [],
      sort_order: p.sort_order,
      translations: tr,
      variants: p.variants?.map((v) => ({ ...v })) ?? [],
      modifier_groups: p.modifier_groups?.map((g) => ({ ...g, modifiers: g.modifiers?.map((m) => ({ ...m })) ?? [] })) ?? [],
    });
    setEditing(p);
    setErrors({});
    setTab("en");
    setOpen(true);
  }

  function validate() {
    const e: Record<string, string> = {};
    if (!form.slug.trim()) e.slug = "Required";
    if (!form.category_id) e.category_id = "Required";
    if (!form.translations.en.name.trim()) e.name_en = "English name required";
    setErrors(e);
    return !Object.keys(e).length;
  }

  const save = useMutation({
    mutationFn: (body: Partial<Product>) =>
      editing ? api.updateProduct(editing.id, body) : api.createProduct(body),
    onSuccess: async () => {
      qc.invalidateQueries({ queryKey: ["products"] });
      try { await api.flushCache(); } catch { /* non-critical */ }
      setOpen(false);
      show(editing ? "Product updated" : "Product created");
    },
    onError: (e) => show(e instanceof Error ? e.message : "Save failed", "error"),
  });

  const del = useMutation({
    mutationFn: (id: number) => api.deleteProduct(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["products"] }); setToDelete(null); show("Deleted"); },
    onError: (e) => show(e instanceof Error ? e.message : "Delete failed", "error"),
  });

  const toggle = useMutation({
    mutationFn: ({ id, v }: { id: number; v: boolean }) => api.toggleProduct(id, v),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["products"] }),
    onError: () => show("Toggle failed", "error"),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;
    const tr: Product["translations"] = {};
    for (const l of LOCALES) {
      if (form.translations[l].name.trim()) tr[l] = { name: form.translations[l].name, description: form.translations[l].description || undefined };
    }
    save.mutate({
      category_id: form.category_id,
      slug: form.slug,
      image_url: form.image_url || undefined,
      pricing_mode: form.pricing_mode,
      base_price_cents: Math.round(form.base_price_cents),
      is_preorder: form.is_preorder,
      preorder_lead_hours: form.is_preorder ? form.preorder_lead_hours : undefined,
      is_available: form.is_available,
      allergens: form.allergens.length ? form.allergens : undefined,
      sort_order: form.sort_order,
      translations: tr,
      variants: form.variants,
      modifier_groups: form.modifier_groups,
    });
  }

  function addVariant() { setForm((f) => ({ ...f, variants: [...f.variants, { ...BLANK_VARIANT }] })); }
  function removeVariant(i: number) { setForm((f) => ({ ...f, variants: f.variants.filter((_, j) => j !== i) })); }
  function setVariant(i: number, patch: Partial<ProductVariant>) {
    setForm((f) => ({ ...f, variants: f.variants.map((v, j) => j === i ? { ...v, ...patch } : v) }));
  }
  function addGroup() { setForm((f) => ({ ...f, modifier_groups: [...f.modifier_groups, { ...BLANK_GROUP, modifiers: [{ ...BLANK_MOD }] }] })); }
  function removeGroup(i: number) { setForm((f) => ({ ...f, modifier_groups: f.modifier_groups.filter((_, j) => j !== i) })); }
  function setGroup(i: number, patch: Partial<ModifierGroup>) {
    setForm((f) => ({ ...f, modifier_groups: f.modifier_groups.map((g, j) => j === i ? { ...g, ...patch } : g) }));
  }
  function addModifier(gi: number) { setGroup(gi, { modifiers: [...(form.modifier_groups[gi]?.modifiers ?? []), { ...BLANK_MOD }] }); }
  function removeModifier(gi: number, mi: number) {
    setGroup(gi, { modifiers: form.modifier_groups[gi].modifiers.filter((_, j) => j !== mi) });
  }
  function setModifier(gi: number, mi: number, patch: Partial<typeof BLANK_MOD>) {
    setGroup(gi, { modifiers: form.modifier_groups[gi].modifiers.map((m, j) => j === mi ? { ...m, ...patch } : m) });
  }

  const categoryName = (id: number) => cats.data?.items.find((c) => c.id === id)?.translations?.en?.name ?? `#${id}`;
  const cents2dollars = (c: number) => (c / 100).toFixed(2);
  const dollars2cents = (v: string) => Math.round(parseFloat(v || "0") * 100);

  return (
    <div>
      <PageHeader
        title="Products"
        action={<button className="btn-primary" onClick={openNew}>+ New product</button>}
      />

      {/* Category filter */}
      <div className="mb-4 flex flex-wrap gap-2">
        <button className={`btn-outline text-xs ${!selectedCat ? "border-brand-500 text-brand-600" : ""}`} onClick={() => setSelectedCat(0)}>All</button>
        {cats.data?.items.map((c) => (
          <button key={c.id} className={`btn-outline text-xs ${selectedCat === c.id ? "border-brand-500 text-brand-600" : ""}`} onClick={() => setSelectedCat(c.id)}>
            {c.translations?.en?.name ?? c.slug}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <div key={i} className="skeleton h-14 rounded-xl" />)}</div>
      ) : !data?.items.length ? (
        <EmptyState icon="🍬" title="No products yet" action={<button className="btn-primary" onClick={openNew}>Create one</button>} />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead><tr><th>Product</th><th>Category</th><th>Price</th><th>Available</th><th></th></tr></thead>
            <tbody>
              {data.items.map((p) => (
                <tr key={p.id}>
                  <td>
                    <div className="flex items-center gap-3">
                      {p.image_url && <img src={p.image_url} alt="" className="h-9 w-9 rounded-lg object-cover" />}
                      <div>
                        <p className="font-medium">{p.translations?.en?.name ?? p.slug}</p>
                        <p className="text-xs text-gray-400 font-mono">{p.slug}</p>
                      </div>
                    </div>
                  </td>
                  <td className="text-xs text-gray-500">{categoryName(p.category_id)}</td>
                  <td>${cents2dollars(p.base_price_cents)}</td>
                  <td>
                    <button
                      className={`switch ${p.is_available ? "bg-brand-500" : "bg-gray-200"}`}
                      onClick={() => toggle.mutate({ id: p.id, v: !p.is_available })}
                    >
                      <span className={`switch-thumb ${p.is_available ? "translate-x-4" : "translate-x-0"}`} />
                    </button>
                  </td>
                  <td>
                    <div className="flex gap-1">
                      <button className="btn-ghost py-1 px-2 text-xs" onClick={() => openEdit(p)}>Edit</button>
                      <button className="btn-ghost py-1 px-2 text-xs text-red-500" onClick={() => setToDelete(p)}>Delete</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Product Form Modal */}
      <Modal open={open} onClose={() => setOpen(false)} title={editing ? "Edit Product" : "New Product"} size="xl">
        <form onSubmit={submit} className="space-y-5 max-h-[70vh] overflow-y-auto pr-1" noValidate>
          {/* Core fields */}
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Category" required error={errors.category_id}>
              <select className="input" value={form.category_id} onChange={(e) => setForm((f) => ({ ...f, category_id: Number(e.target.value) }))}>
                <option value={0}>Select…</option>
                {cats.data?.items.map((c) => <option key={c.id} value={c.id}>{c.translations?.en?.name ?? c.slug}</option>)}
              </select>
            </FormField>
            <FormField label="Slug" required error={errors.slug}>
              <input className="input" value={form.slug} onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value }))} />
            </FormField>
            <FormField label="Pricing mode">
              <select className="input" value={form.pricing_mode} onChange={(e) => setForm((f) => ({ ...f, pricing_mode: e.target.value as "unit" | "weight" }))}>
                <option value="unit">Unit (fixed price)</option>
                <option value="weight">Weight (per gram)</option>
              </select>
            </FormField>
            <FormField label="Base price (CAD)">
              <input type="number" min={0} step={0.01} className="input" value={cents2dollars(form.base_price_cents)} onChange={(e) => setForm((f) => ({ ...f, base_price_cents: dollars2cents(e.target.value) }))} />
            </FormField>
            <FormField label="Sort order">
              <input type="number" className="input" value={form.sort_order} onChange={(e) => setForm((f) => ({ ...f, sort_order: Number(e.target.value) }))} />
            </FormField>
          </div>

          <div className="flex gap-6">
            {[
              { label: "Available", key: "is_available" as const },
              { label: "Pre-order", key: "is_preorder" as const },
            ].map(({ label, key }) => (
              <label key={key} className="flex cursor-pointer items-center gap-2 text-sm">
                <button type="button" className={`switch ${form[key] ? "bg-brand-500" : "bg-gray-200"}`} onClick={() => setForm((f) => ({ ...f, [key]: !f[key] }))}>
                  <span className={`switch-thumb ${form[key] ? "translate-x-4" : "translate-x-0"}`} />
                </button>
                {label}
              </label>
            ))}
          </div>

          <MediaPicker value={form.image_url} onChange={(url) => setForm((f) => ({ ...f, image_url: url }))} />

          {/* Translations */}
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
                <input className="input" value={form.translations[tab].name} onChange={(e) => setForm((f) => ({ ...f, translations: { ...f.translations, [tab]: { ...f.translations[tab], name: e.target.value } } }))} />
              </FormField>
              <FormField label="Description">
                <textarea rows={2} className="input resize-none" value={form.translations[tab].description ?? ""} onChange={(e) => setForm((f) => ({ ...f, translations: { ...f.translations, [tab]: { ...f.translations[tab], description: e.target.value } } }))} />
              </FormField>
            </div>
          </div>

          {/* Variants */}
          <div>
            <div className="mb-2 flex items-center justify-between">
              <h3 className="text-sm font-semibold">Variants</h3>
              <button type="button" className="btn-ghost text-xs" onClick={addVariant}>+ Add variant</button>
            </div>
            {form.variants.map((v, i) => (
              <div key={i} className="mb-2 grid grid-cols-4 gap-2 rounded-lg bg-gray-50 p-3">
                <FormField label="Label"><input className="input text-xs" value={v.label} onChange={(e) => setVariant(i, { label: e.target.value })} /></FormField>
                <FormField label="Price ($)"><input type="number" min={0} step={0.01} className="input text-xs" value={cents2dollars(v.price_cents)} onChange={(e) => setVariant(i, { price_cents: dollars2cents(e.target.value) })} /></FormField>
                <FormField label="SKU"><input className="input text-xs" value={v.sku ?? ""} onChange={(e) => setVariant(i, { sku: e.target.value })} /></FormField>
                <div className="flex items-end gap-2">
                  <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={v.is_default} onChange={(e) => setVariant(i, { is_default: e.target.checked })} /> Default</label>
                  <button type="button" className="text-red-400 hover:text-red-600 text-xs ml-auto" onClick={() => removeVariant(i)}>✕</button>
                </div>
              </div>
            ))}
          </div>

          {/* Modifier groups */}
          <div>
            <div className="mb-2 flex items-center justify-between">
              <h3 className="text-sm font-semibold">Modifier Groups</h3>
              <button type="button" className="btn-ghost text-xs" onClick={addGroup}>+ Add group</button>
            </div>
            {form.modifier_groups.map((g, gi) => (
              <div key={gi} className="mb-3 rounded-xl border border-gray-200 p-3">
                <div className="mb-2 grid grid-cols-3 gap-2">
                  <FormField label="Group label"><input className="input text-xs" value={g.label} onChange={(e) => setGroup(gi, { label: e.target.value })} /></FormField>
                  <FormField label="Min"><input type="number" min={0} className="input text-xs" value={g.min_select} onChange={(e) => setGroup(gi, { min_select: Number(e.target.value) })} /></FormField>
                  <FormField label="Max"><input type="number" min={1} className="input text-xs" value={g.max_select} onChange={(e) => setGroup(gi, { max_select: Number(e.target.value) })} /></FormField>
                </div>
                <div className="mb-2 flex gap-3 text-xs">
                  <label className="flex items-center gap-1"><input type="checkbox" checked={g.is_required} onChange={(e) => setGroup(gi, { is_required: e.target.checked })} /> Required</label>
                  <button type="button" className="ml-auto text-red-400" onClick={() => removeGroup(gi)}>Remove group</button>
                </div>
                {g.modifiers.map((m, mi) => (
                  <div key={mi} className="mb-1 flex gap-2 rounded bg-gray-50 p-2">
                    <input className="input text-xs flex-1" placeholder="Option label" value={m.label} onChange={(e) => setModifier(gi, mi, { label: e.target.value })} />
                    <input type="number" step={0.01} className="input text-xs w-20" placeholder="$Δ" value={cents2dollars(m.price_delta_cents)} onChange={(e) => setModifier(gi, mi, { price_delta_cents: dollars2cents(e.target.value) })} />
                    <button type="button" className="text-red-400 text-xs" onClick={() => removeModifier(gi, mi)}>✕</button>
                  </div>
                ))}
                <button type="button" className="mt-1 text-xs text-brand-600 hover:underline" onClick={() => addModifier(gi)}>+ option</button>
              </div>
            ))}
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" className="btn-ghost" onClick={() => setOpen(false)}>Cancel</button>
            <button type="submit" className="btn-primary" disabled={save.isPending}>{save.isPending ? "Saving…" : "Save"}</button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!toDelete}
        title="Delete product?"
        message={`Delete "${toDelete?.translations?.en?.name ?? toDelete?.slug}"?`}
        confirmLabel="Delete"
        danger
        loading={del.isPending}
        onConfirm={() => toDelete && del.mutate(toDelete.id)}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { Modal } from "@/components/Modal";
import { FormField } from "@/components/FormField";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import type { Language } from "@/api/types";

function blankForm(): Partial<Language> {
  return { code: "", name: "", native_name: "", is_rtl: false, is_active: true, sort_order: 0 };
}

export function LanguagesPage() {
  const [form, setForm] = useState(blankForm());
  const [editing, setEditing] = useState<Language | null>(null);
  const [open, setOpen] = useState(false);
  const [toDelete, setToDelete] = useState<Language | null>(null);
  const [bundleLocale, setBundleLocale] = useState<string | null>(null);
  const [bundle, setBundle] = useState<Record<string, string>>({});
  const { show } = useToast();
  const qc = useQueryClient();

  const { data: langs = [], isLoading } = useQuery({ queryKey: qk.languages(), queryFn: api.languages });

  const bundleQ = useQuery({
    queryKey: qk.uiBundle(bundleLocale ?? ""),
    queryFn: () => api.uiBundle(bundleLocale!),
    enabled: !!bundleLocale,
  });

  function openNew() { setForm(blankForm()); setEditing(null); setOpen(true); }
  function openEdit(l: Language) { setForm({ ...l }); setEditing(l); setOpen(true); }
  function openBundle(locale: string) {
    setBundleLocale(locale);
    setBundle(bundleQ.data ?? {});
  }

  const save = useMutation({
    mutationFn: (body: Partial<Language>) => editing ? api.updateLanguage(editing.id, body) : api.createLanguage(body),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.languages() }); setOpen(false); show(editing ? "Updated" : "Created"); },
    onError: (e) => show(e instanceof Error ? e.message : "Save failed", "error"),
  });

  const del = useMutation({
    mutationFn: (id: number) => api.deleteLanguage(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: qk.languages() }); setToDelete(null); show("Deleted"); },
    onError: () => show("Delete failed", "error"),
  });

  const saveBundle = useMutation({
    mutationFn: () => api.saveUiBundle(bundleLocale!, bundle),
    onSuccess: () => { setBundleLocale(null); show("Bundle saved"); },
    onError: () => show("Save failed", "error"),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    save.mutate(form);
  }

  // Sync bundle when query loads
  if (bundleQ.data && Object.keys(bundle).length === 0 && bundleLocale) {
    setBundle(bundleQ.data);
  }

  return (
    <div>
      <PageHeader title="Languages" action={<button className="btn-primary" onClick={openNew}>+ New language</button>} />

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-12 rounded-xl" />)}</div>
      ) : !langs.length ? (
        <EmptyState icon="🌐" title="No languages yet" action={<button className="btn-primary" onClick={openNew}>Add language</button>} />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead><tr><th>Code</th><th>Name</th><th>Native</th><th>RTL</th><th>Default</th><th>Active</th><th></th></tr></thead>
            <tbody>
              {langs.map((l) => (
                <tr key={l.id}>
                  <td className="font-mono text-xs font-bold">{l.code}</td>
                  <td>{l.name}</td>
                  <td>{l.native_name}</td>
                  <td>{l.is_rtl ? "Yes" : "No"}</td>
                  <td>{l.is_default && <span className="badge badge-blue">Default</span>}</td>
                  <td><span className={`badge ${l.is_active ? "badge-green" : "badge-gray"}`}>{l.is_active ? "On" : "Off"}</span></td>
                  <td>
                    <div className="flex gap-1">
                      <button className="btn-ghost py-1 px-2 text-xs" onClick={() => openEdit(l)}>Edit</button>
                      <button className="btn-ghost py-1 px-2 text-xs text-brand-600" onClick={() => openBundle(l.code)}>Strings</button>
                      <button className="btn-ghost py-1 px-2 text-xs text-red-500" onClick={() => setToDelete(l)}>Delete</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Language form */}
      <Modal open={open} onClose={() => setOpen(false)} title={editing ? "Edit Language" : "New Language"} size="sm">
        <form onSubmit={submit} className="space-y-4" noValidate>
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Code" required><input className="input uppercase" maxLength={5} value={form.code} onChange={(e) => setForm((f) => ({ ...f, code: e.target.value.toLowerCase() }))} placeholder="en" /></FormField>
            <FormField label="Sort order"><input type="number" className="input" value={form.sort_order ?? 0} onChange={(e) => setForm((f) => ({ ...f, sort_order: Number(e.target.value) }))} /></FormField>
            <FormField label="Name" required><input className="input" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} placeholder="English" /></FormField>
            <FormField label="Native name"><input className="input" value={form.native_name} onChange={(e) => setForm((f) => ({ ...f, native_name: e.target.value }))} placeholder="English" /></FormField>
          </div>
          <div className="flex gap-6">
            {[
              { label: "Active", key: "is_active" as const },
              { label: "RTL", key: "is_rtl" as const },
              { label: "Default", key: "is_default" as const },
            ].map(({ label, key }) => (
              <label key={key} className="flex cursor-pointer items-center gap-2 text-sm">
                <button type="button" className={`switch ${form[key] ? "bg-brand-500" : "bg-gray-200"}`} onClick={() => setForm((f) => ({ ...f, [key]: !f[key] }))}>
                  <span className={`switch-thumb ${form[key] ? "translate-x-4" : "translate-x-0"}`} />
                </button>
                {label}
              </label>
            ))}
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" className="btn-ghost" onClick={() => setOpen(false)}>Cancel</button>
            <button type="submit" className="btn-primary" disabled={save.isPending}>{save.isPending ? "Saving…" : "Save"}</button>
          </div>
        </form>
      </Modal>

      {/* UI Strings bundle editor */}
      <Modal open={!!bundleLocale} onClose={() => setBundleLocale(null)} title={`UI Strings — ${bundleLocale?.toUpperCase()}`} size="xl">
        {bundleQ.isLoading ? (
          <div className="space-y-2">{Array.from({ length: 8 }).map((_, i) => <div key={i} className="skeleton h-10 rounded" />)}</div>
        ) : (
          <div className="space-y-3 max-h-[60vh] overflow-y-auto">
            {Object.keys(bundleQ.data ?? {}).map((key) => (
              <div key={key} className="grid grid-cols-2 gap-3 items-center">
                <p className="text-xs font-mono text-gray-500 truncate">{key}</p>
                <input
                  className="input text-xs"
                  value={bundle[key] ?? ""}
                  onChange={(e) => setBundle((b) => ({ ...b, [key]: e.target.value }))}
                />
              </div>
            ))}
          </div>
        )}
        <div className="mt-4 flex justify-end gap-2">
          <button className="btn-ghost" onClick={() => setBundleLocale(null)}>Cancel</button>
          <button className="btn-primary" onClick={() => saveBundle.mutate()} disabled={saveBundle.isPending}>{saveBundle.isPending ? "Saving…" : "Save bundle"}</button>
        </div>
      </Modal>

      <ConfirmDialog
        open={!!toDelete} title="Delete language?" message={`Delete "${toDelete?.name}"?`}
        confirmLabel="Delete" danger loading={del.isPending}
        onConfirm={() => toDelete && del.mutate(toDelete.id)}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}

import { useEffect, useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { PageHeader } from "@/components/PageHeader";
import { FormField } from "@/components/FormField";
import type { Settings } from "@/api/types";

type FormState = Omit<Settings, "tax_percent"> & { tax_str: string };

function toForm(s: Settings): FormState {
  return { ...s, tax_str: String(s.tax_percent ?? 0) };
}

export function SettingsPage() {
  const [form, setForm] = useState<FormState | null>(null);
  const { show } = useToast();

  const { data, isLoading } = useQuery({ queryKey: ["settings"], queryFn: api.getSettings });

  useEffect(() => { if (data && !form) setForm(toForm(data)); }, [data, form]);

  const save = useMutation({
    mutationFn: (f: FormState) => api.saveSettings({ ...f, tax_percent: parseFloat(f.tax_str || "0") }),
    onSuccess: () => show("Settings saved"),
    onError: (e) => show(e instanceof Error ? e.message : "Save failed", "error"),
  });

  if (isLoading || !form) {
    return <div className="space-y-4">{Array.from({ length: 6 }).map((_, i) => <div key={i} className="skeleton h-10 rounded-xl" />)}</div>;
  }

  function set<K extends keyof FormState>(k: K, v: FormState[K]) { setForm((f) => f ? { ...f, [k]: v } : f); }

  return (
    <div>
      <PageHeader title="Settings" description="Store configuration and integrations." />

      <form
        onSubmit={(e) => { e.preventDefault(); if (form) save.mutate(form); }}
        className="max-w-2xl space-y-8"
      >
        {/* Payments */}
        <section className="card p-5 space-y-4">
          <h2 className="font-semibold text-gray-800">Payments</h2>
          <div className="flex gap-6">
            {[
              { label: "Cash on delivery", key: "cod_enabled" as const },
              { label: "Square", key: "square_enabled" as const },
            ].map(({ label, key }) => (
              <label key={key} className="flex cursor-pointer items-center gap-2 text-sm">
                <button type="button" className={`switch ${form[key] ? "bg-brand-500" : "bg-gray-200"}`} onClick={() => set(key, !form[key])}>
                  <span className={`switch-thumb ${form[key] ? "translate-x-4" : "translate-x-0"}`} />
                </button>
                {label}
              </label>
            ))}
          </div>
          {form.square_enabled && (
            <div className="grid grid-cols-2 gap-4">
              <FormField label="Environment">
                <select className="input" value={form.square_environment} onChange={(e) => set("square_environment", e.target.value)}>
                  <option value="sandbox">Sandbox</option>
                  <option value="production">Production</option>
                </select>
              </FormField>
              <FormField label="Application ID">
                <input className="input" value={form.square_application_id ?? ""} onChange={(e) => set("square_application_id", e.target.value)} />
              </FormField>
              <FormField label="Access token">
                <input type="password" className="input" value={form.square_access_token ?? ""} onChange={(e) => set("square_access_token", e.target.value)} placeholder="••••••••" />
              </FormField>
              <FormField label="Location ID">
                <input className="input" value={form.square_location_id ?? ""} onChange={(e) => set("square_location_id", e.target.value)} />
              </FormField>
            </div>
          )}
        </section>

        {/* Store */}
        <section className="card p-5 space-y-4">
          <h2 className="font-semibold text-gray-800">Store</h2>
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Currency">
              <input className="input" value={form.currency} onChange={(e) => set("currency", e.target.value)} placeholder="CAD" />
            </FormField>
            <FormField label="Tax %" hint="e.g. 13 for 13%">
              <input type="number" min={0} max={100} step={0.01} className="input" value={form.tax_str} onChange={(e) => set("tax_str", e.target.value)} />
            </FormField>
          </div>
        </section>

        {/* Meta Pixel */}
        <section className="card p-5 space-y-4">
          <h2 className="font-semibold text-gray-800">Meta Pixel / CAPI</h2>
          <div className="grid grid-cols-2 gap-4">
            <FormField label="Pixel ID">
              <input className="input" value={form.meta_pixel_id ?? ""} onChange={(e) => set("meta_pixel_id", e.target.value)} />
            </FormField>
            <FormField label="Test event code" hint="For CAPI testing">
              <input className="input" value={form.meta_test_event_code ?? ""} onChange={(e) => set("meta_test_event_code", e.target.value)} />
            </FormField>
            <FormField label="CAPI token" hint="Server-side event token">
              <input type="password" className="input" value={form.meta_capi_token ?? ""} onChange={(e) => set("meta_capi_token", e.target.value)} placeholder="••••••••" />
            </FormField>
          </div>
        </section>

        <div className="flex justify-end">
          <button type="submit" className="btn-primary" disabled={save.isPending}>{save.isPending ? "Saving…" : "Save settings"}</button>
        </div>
      </form>
    </div>
  );
}

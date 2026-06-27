import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { api, qk } from "@/api/endpoints";
import { CA_PROVINCES } from "@/lib/provinces";
import { vRequired, vPhone, vPostalCA } from "@/lib/validate";
import type { Address } from "@/api/types";

type FormState = {
  label: string; line1: string; line2: string;
  city: string; province_code: string; postal_code: string; phone: string;
  lat?: number; lng?: number;
};

type Fields = "line1" | "city" | "postal_code" | "phone";
type Errors = Partial<Record<Fields, string>>;

const BLANK: FormState = {
  label: "", line1: "", line2: "", city: "",
  province_code: "ON", postal_code: "", phone: "",
};

function runValidation(f: FormState): Errors {
  return {
    line1:       vRequired(f.line1)      ?? undefined,
    city:        vRequired(f.city)       ?? undefined,
    postal_code: vPostalCA(f.postal_code) ?? undefined,
    phone:       vPhone(f.phone)         ?? undefined,
  };
}

function FieldError({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <p className="mt-1 text-xs text-red-500">{msg}</p>;
}

export function AddressesPage() {
  const { t } = useTranslation();
  const qc = useQueryClient();

  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<FormState>(BLANK);
  const [errors, setErrors] = useState<Errors>({});
  const [touched, setTouched] = useState<Set<Fields>>(new Set());
  const [locating, setLocating] = useState(false);
  const [serverError, setServerError] = useState("");

  const { data: addresses = [], isLoading } = useQuery({
    queryKey: qk.addresses(),
    queryFn: api.addresses,
  });

  const invalidate = () => qc.invalidateQueries({ queryKey: qk.addresses() });

  const addMutation = useMutation({
    mutationFn: (body: Partial<Address>) => api.addAddress(body),
    onSuccess: () => {
      invalidate();
      setShowForm(false);
      setForm(BLANK);
      setErrors({});
      setTouched(new Set());
      setServerError("");
    },
    onError: (e) => setServerError(e instanceof Error ? e.message : t("common.error")),
  });

  const deleteMutation  = useMutation({ mutationFn: api.deleteAddress,       onSuccess: invalidate });
  const defaultMutation = useMutation({ mutationFn: api.setDefaultAddress,   onSuccess: invalidate });

  function touch(field: Fields) {
    setTouched((prev) => new Set(prev).add(field));
  }

  const validators: Record<Fields, (v: string) => string | null> = {
    line1: vRequired, city: vRequired, postal_code: vPostalCA, phone: vPhone,
  };

  function handleChange(field: keyof FormState) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
      const val = e.target.value;
      setForm((f) => ({ ...f, [field]: val }));
      if (field in validators && touched.has(field as Fields)) {
        setErrors((prev) => ({ ...prev, [field]: (validators[field as Fields](val) ?? undefined) }));
      }
    };
  }

  function handleBlur(field: Fields) {
    touch(field);
    setErrors((prev) => ({ ...prev, [field]: validators[field](form[field] as string) ?? undefined }));
  }

  function inputClass(field: Fields) {
    return `input ${touched.has(field) && errors[field] ? "border-red-400 focus:border-red-400 focus:ring-red-100" : ""}`;
  }

  function locate() {
    if (!navigator.geolocation) return;
    setLocating(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => { setForm((f) => ({ ...f, lat: pos.coords.latitude, lng: pos.coords.longitude })); setLocating(false); },
      () => setLocating(false),
    );
  }

  function submit(e: React.FormEvent) {
    e.preventDefault();
    const errs = runValidation(form);
    setErrors(errs);
    setTouched(new Set<Fields>(["line1", "city", "postal_code", "phone"]));
    if (Object.values(errs).some(Boolean)) return;

    addMutation.mutate({
      label: form.label || undefined,
      line1: form.line1,
      line2: form.line2 || undefined,
      city: form.city,
      province_code: form.province_code,
      postal_code: form.postal_code,
      phone: form.phone || undefined,
      lat: form.lat,
      lng: form.lng,
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("address.title")}</h1>
        {!showForm && (
          <button className="btn-primary" onClick={() => setShowForm(true)}>
            {t("address.add")}
          </button>
        )}
      </div>

      {showForm && (
        <form onSubmit={submit} className="card space-y-3 p-4" noValidate>
          <h2 className="font-semibold">{t("address.add")}</h2>

          {/* Label — optional, no validation */}
          <input
            className="input"
            placeholder={t("address.label")}
            value={form.label}
            onChange={handleChange("label")}
          />

          {/* Line 1 */}
          <div>
            <input
              className={inputClass("line1")}
              placeholder={t("address.line1")}
              value={form.line1}
              onChange={handleChange("line1")}
              onBlur={() => handleBlur("line1")}
            />
            <FieldError msg={touched.has("line1") && errors.line1 ? t(errors.line1) : undefined} />
          </div>

          {/* Line 2 — optional */}
          <input
            className="input"
            placeholder={t("address.line2")}
            value={form.line2}
            onChange={handleChange("line2")}
          />

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {/* City */}
            <div>
              <input
                className={inputClass("city")}
                placeholder={t("address.city")}
                value={form.city}
                onChange={handleChange("city")}
                onBlur={() => handleBlur("city")}
              />
              <FieldError msg={touched.has("city") && errors.city ? t(errors.city) : undefined} />
            </div>

            {/* Province — always valid, no validation needed */}
            <select className="input" value={form.province_code} onChange={handleChange("province_code")}>
              {CA_PROVINCES.map((p) => (
                <option key={p.code} value={p.code}>{p.name} ({p.code})</option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {/* Postal code */}
            <div>
              <input
                className={inputClass("postal_code")}
                placeholder={t("address.postalCode")}
                value={form.postal_code}
                onChange={handleChange("postal_code")}
                onBlur={() => handleBlur("postal_code")}
              />
              <FieldError msg={touched.has("postal_code") && errors.postal_code ? t(errors.postal_code) : undefined} />
            </div>

            {/* Phone */}
            <div>
              <input
                className={inputClass("phone")}
                type="tel"
                placeholder={t("address.phone")}
                value={form.phone}
                onChange={handleChange("phone")}
                onBlur={() => handleBlur("phone")}
              />
              <FieldError msg={touched.has("phone") && errors.phone ? t(errors.phone) : undefined} />
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button type="button" className="btn-ghost text-sm" onClick={locate} disabled={locating}>
              {locating ? t("address.locating") : t("address.useLocation")}
            </button>
            {form.lat != null && (
              <span className="text-xs text-brand-600">
                {form.lat.toFixed(5)}, {form.lng?.toFixed(5)}
              </span>
            )}
          </div>

          {serverError && <p className="text-sm text-red-600">{serverError}</p>}

          <div className="flex gap-2">
            <button type="submit" className="btn-primary" disabled={addMutation.isPending}>
              {addMutation.isPending ? t("common.loading") : t("address.save")}
            </button>
            <button
              type="button"
              className="btn-ghost"
              onClick={() => { setShowForm(false); setForm(BLANK); setErrors({}); setTouched(new Set()); setServerError(""); }}
            >
              {t("address.cancel")}
            </button>
          </div>
        </form>
      )}

      {isLoading ? (
        <p className="text-brand-700/60">{t("common.loading")}</p>
      ) : addresses.length === 0 ? (
        <div className="rounded-2xl border-2 border-dashed border-brand-200 p-8 text-center">
          <p className="text-brand-700/60">{t("address.empty")}</p>
          {!showForm && (
            <button className="btn-primary mt-4" onClick={() => setShowForm(true)}>
              {t("address.add")}
            </button>
          )}
        </div>
      ) : (
        <ul className="space-y-3">
          {addresses.map((a) => (
            <li key={a.id} className="card flex items-start justify-between gap-3 p-4">
              <div className="min-w-0 space-y-0.5">
                <div className="flex flex-wrap items-center gap-2">
                  {a.label && <span className="font-semibold">{a.label}</span>}
                  {a.is_default && (
                    <span className="rounded-full bg-brand-100 px-2 py-0.5 text-xs font-medium text-brand-700">
                      {t("address.default")}
                    </span>
                  )}
                </div>
                <p className="truncate text-sm text-gray-700">
                  {a.line1}{a.line2 ? `, ${a.line2}` : ""}
                </p>
                <p className="text-sm text-gray-500">
                  {a.city}, {a.province_code} {a.postal_code}
                </p>
                {a.phone && <p className="text-xs text-gray-400">{a.phone}</p>}
              </div>

              <div className="flex shrink-0 flex-col gap-1">
                {!a.is_default && (
                  <button
                    className="btn-ghost py-1 text-xs"
                    onClick={() => defaultMutation.mutate(a.id)}
                    disabled={defaultMutation.isPending}
                  >
                    {t("address.setDefault")}
                  </button>
                )}
                <button
                  className="btn-ghost py-1 text-xs text-red-600 hover:bg-red-50"
                  onClick={() => deleteMutation.mutate(a.id)}
                  disabled={deleteMutation.isPending}
                >
                  {t("address.delete")}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}

      <p className="text-center text-sm text-brand-700/50">
        <Link to="/checkout" className="underline hover:text-brand-700">
          ← {t("cart.checkout")}
        </Link>
      </p>
    </div>
  );
}

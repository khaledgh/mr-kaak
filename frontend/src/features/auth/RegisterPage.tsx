import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { api } from "@/api/endpoints";
import { extractApiError } from "@/api/client";
import { useAuth } from "@/stores/auth";
import { vEmail, vName, vPassword, vPhone } from "@/lib/validate";

type Fields = "name" | "email" | "phone" | "password";
type Errors = Partial<Record<Fields, string>>;

function runValidation(form: Record<Fields, string>): Errors {
  return {
    name:     vName(form.name)         ?? undefined,
    email:    vEmail(form.email)       ?? undefined,
    phone:    vPhone(form.phone)       ?? undefined,
    password: vPassword(form.password) ?? undefined,
  };
}

function FieldError({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <p className="mt-1 text-xs text-red-500">{msg}</p>;
}

export function RegisterPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const setAuth = useAuth((s) => s.setAuth);

  const [form, setForm] = useState<Record<Fields, string>>({ name: "", email: "", phone: "", password: "" });
  const [errors, setErrors] = useState<Errors>({});
  const [touched, setTouched] = useState<Set<Fields>>(new Set());
  const [serverError, setServerError] = useState("");
  const [busy, setBusy] = useState(false);

  function touch(field: Fields) {
    setTouched((prev) => new Set(prev).add(field));
  }

  function handleChange(field: Fields) {
    return (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value;
      setForm((f) => ({ ...f, [field]: val }));
      if (touched.has(field)) {
        const validators: Record<Fields, (v: string) => string | null> = {
          name: vName, email: vEmail, phone: vPhone, password: vPassword,
        };
        setErrors((prev) => ({ ...prev, [field]: validators[field](val) ?? undefined }));
      }
    };
  }

  function handleBlur(field: Fields) {
    touch(field);
    const validators: Record<Fields, (v: string) => string | null> = {
      name: vName, email: vEmail, phone: vPhone, password: vPassword,
    };
    setErrors((prev) => ({ ...prev, [field]: validators[field](form[field]) ?? undefined }));
  }

  function inputClass(field: Fields) {
    return `input ${touched.has(field) && errors[field] ? "border-red-400 focus:border-red-400 focus:ring-red-100" : ""}`;
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const errs = runValidation(form);
    setErrors(errs);
    setTouched(new Set<Fields>(["name", "email", "phone", "password"]));
    if (Object.values(errs).some(Boolean)) return;

    setServerError("");
    setBusy(true);
    try {
      const res = await api.register({
        name: form.name,
        email: form.email,
        password: form.password,
        phone: form.phone || undefined,
      });
      setAuth(res.user, res.tokens);
      navigate("/");
    } catch (err) {
      setServerError(extractApiError(err, t("common.error")));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="mx-auto max-w-sm space-y-6 animate-toast" noValidate>
      <div className="flex flex-col items-center text-center">
        <Link to="/" className="mb-4 hover:opacity-90 transition-opacity">
          <img src="/logo.png" className="h-16 w-auto object-contain" alt={t("app.title")} />
        </Link>
        <h1 className="text-2xl font-bold text-brand-900">{t("auth.register")}</h1>
      </div>

      <div>
        <input
          className={inputClass("name")}
          placeholder={t("auth.name")}
          value={form.name}
          autoComplete="name"
          onChange={handleChange("name")}
          onBlur={() => handleBlur("name")}
        />
        <FieldError msg={touched.has("name") && errors.name ? t(errors.name) : undefined} />
      </div>

      <div>
        <input
          className={inputClass("email")}
          type="email"
          placeholder={t("auth.email")}
          value={form.email}
          autoComplete="email"
          onChange={handleChange("email")}
          onBlur={() => handleBlur("email")}
        />
        <FieldError msg={touched.has("email") && errors.email ? t(errors.email) : undefined} />
      </div>

      <div>
        <input
          className={inputClass("phone")}
          type="tel"
          placeholder={t("auth.phone")}
          value={form.phone}
          autoComplete="tel"
          onChange={handleChange("phone")}
          onBlur={() => handleBlur("phone")}
        />
        <FieldError msg={touched.has("phone") && errors.phone ? t(errors.phone) : undefined} />
      </div>

      <div>
        <input
          className={inputClass("password")}
          type="password"
          placeholder={t("auth.password")}
          value={form.password}
          autoComplete="new-password"
          onChange={handleChange("password")}
          onBlur={() => handleBlur("password")}
        />
        <FieldError msg={touched.has("password") && errors.password ? t(errors.password) : undefined} />
      </div>

      {serverError && <p className="text-sm text-red-600">{serverError}</p>}

      <button className="btn-primary w-full" disabled={busy}>
        {busy ? t("common.loading") : t("auth.register")}
      </button>

      <p className="text-center text-sm">
        {t("auth.haveAccount")}{" "}
        <Link to="/login" className="text-brand-700 underline">{t("auth.login")}</Link>
      </p>
    </form>
  );
}

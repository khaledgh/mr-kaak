import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { api } from "@/api/endpoints";
import { extractApiError } from "@/api/client";
import { useAuth } from "@/stores/auth";
import { vEmail, vPassword } from "@/lib/validate";

type Errors = { email?: string; password?: string };

function runValidation(email: string, password: string): Errors {
  return {
    email: vEmail(email) ?? undefined,
    password: vPassword(password) ?? undefined,
  };
}

function FieldError({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <p className="mt-1 text-xs text-red-500">{msg}</p>;
}

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const setAuth = useAuth((s) => s.setAuth);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<Errors>({});
  const [touched, setTouched] = useState<Set<string>>(new Set());
  const [serverError, setServerError] = useState("");
  const [busy, setBusy] = useState(false);

  function touch(field: string) {
    setTouched((prev) => new Set(prev).add(field));
  }

  function fieldError(field: keyof Errors) {
    return touched.has(field) ? t(errors[field] ?? "") || undefined : undefined;
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const errs = runValidation(email, password);
    setErrors(errs);
    setTouched(new Set(["email", "password"]));
    if (errs.email || errs.password) return;

    setServerError("");
    setBusy(true);
    try {
      const res = await api.login({ email, password });
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
        <h1 className="text-2xl font-bold text-brand-900">{t("auth.login")}</h1>
      </div>

      <div>
        <input
          className={`input ${touched.has("email") && errors.email ? "border-red-400 focus:border-red-400 focus:ring-red-100" : ""}`}
          type="email"
          placeholder={t("auth.email")}
          value={email}
          autoComplete="email"
          onChange={(e) => {
            setEmail(e.target.value);
            if (touched.has("email")) setErrors((prev) => ({ ...prev, email: vEmail(e.target.value) ?? undefined }));
          }}
          onBlur={() => { touch("email"); setErrors((prev) => ({ ...prev, email: vEmail(email) ?? undefined })); }}
        />
        <FieldError msg={fieldError("email") ? t(errors.email!) : undefined} />
      </div>

      <div>
        <input
          className={`input ${touched.has("password") && errors.password ? "border-red-400 focus:border-red-400 focus:ring-red-100" : ""}`}
          type="password"
          placeholder={t("auth.password")}
          value={password}
          autoComplete="current-password"
          onChange={(e) => {
            setPassword(e.target.value);
            if (touched.has("password")) setErrors((prev) => ({ ...prev, password: vPassword(e.target.value) ?? undefined }));
          }}
          onBlur={() => { touch("password"); setErrors((prev) => ({ ...prev, password: vPassword(password) ?? undefined })); }}
        />
        <FieldError msg={fieldError("password") ? t(errors.password!) : undefined} />
      </div>

      {serverError && <p className="text-sm text-red-600">{serverError}</p>}

      <button className="btn-primary w-full" disabled={busy}>
        {busy ? t("common.loading") : t("auth.login")}
      </button>

      <p className="text-center text-sm">
        {t("auth.noAccount")}{" "}
        <Link to="/register" className="text-brand-700 underline">{t("auth.register")}</Link>
      </p>
    </form>
  );
}

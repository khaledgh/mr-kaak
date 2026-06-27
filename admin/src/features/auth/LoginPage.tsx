import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "@/api/endpoints";
import { useAuth } from "@/stores/auth";

const STAFF_ROLES = ["kitchen", "staff", "admin", "super_admin"];

export function LoginPage() {
  const navigate = useNavigate();
  const setAuth = useAuth((s) => s.setAuth);
  const [form, setForm] = useState({ email: "", password: "" });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const set = (k: keyof typeof form) =>
    (e: React.ChangeEvent<HTMLInputElement>) => setForm({ ...form, [k]: e.target.value });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await api.login(form);
      if (!STAFF_ROLES.includes(res.user.role)) {
        setError("Access denied. Staff accounts only.");
        return;
      }
      setAuth(res.user, res.tokens);
      navigate("/orders");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <form onSubmit={submit} className="card w-full max-w-sm space-y-4 p-8">
        <h1 className="text-2xl font-bold text-brand-700">mrkaak Admin</h1>
        <p className="text-sm text-gray-500">Staff sign-in</p>

        <input
          className="input"
          type="email"
          placeholder="Email"
          value={form.email}
          onChange={set("email")}
          required
          autoComplete="email"
        />
        <input
          className="input"
          type="password"
          placeholder="Password"
          value={form.password}
          onChange={set("password")}
          required
          autoComplete="current-password"
        />

        {error && <p className="rounded bg-red-50 px-3 py-2 text-sm text-red-600">{error}</p>}

        <button className="btn-primary w-full" disabled={busy}>
          {busy ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}

import axios from "axios";
import { useAuth } from "@/stores/auth";
import type { Envelope, TokenPair } from "./types";

// Single axios instance. Requests carry the access token and the chosen locale;
// a 401 triggers a one-shot refresh using the refresh token.
export const http = axios.create({ baseURL: import.meta.env.VITE_API_BASE_URL ?? "/api/v1" });

http.interceptors.request.use((config) => {
  const { accessToken } = useAuth.getState();
  if (accessToken) config.headers.Authorization = `Bearer ${accessToken}`;
  const lang = localStorage.getItem("lang");
  if (lang) config.headers["Accept-Language"] = lang;
  return config;
});

let refreshing: Promise<string | null> | null = null;

http.interceptors.response.use(
  (r) => r,
  async (error) => {
    const original = error.config;
    const status = error.response?.status;
    const { refreshToken, setTokens, logout } = useAuth.getState();

    if (status === 401 && refreshToken && !original._retried) {
      original._retried = true;
      refreshing ??= refreshAccess(refreshToken);
      const newToken = await refreshing;
      refreshing = null;
      if (newToken) {
        original.headers.Authorization = `Bearer ${newToken}`;
        return http(original);
      }
      logout();
    }
    return Promise.reject(error);

    async function refreshAccess(token: string): Promise<string | null> {
      try {
        const { data } = await axios.post<Envelope<{ tokens: TokenPair }>>(
          "/api/v1/auth/refresh",
          { refresh_token: token },
        );
        if (data.data?.tokens) {
          setTokens(data.data.tokens);
          return data.data.tokens.access_token;
        }
      } catch {
        /* fall through to logout */
      }
      return null;
    }
  },
);

// unwrap returns the envelope's data or throws the API error message.
export function unwrap<T>(env: Envelope<T>): T {
  if (env.error) throw new Error(env.error.message);
  return env.data as T;
}

// extractApiError reads the structured error message from the API response
// envelope ({"error":{"code":"...","message":"..."}}). Falls back to the raw
// Error message or a generic string if the response has no structured body.
export function extractApiError(err: unknown, fallback = "Something went wrong"): string {
  if (err && typeof err === "object") {
    // Axios error — check response.data.error.message
    const axiosErr = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
    const apiMsg = axiosErr.response?.data?.error?.message;
    if (apiMsg) return apiMsg;
    // Plain Error object
    if (axiosErr.message) return axiosErr.message;
  }
  return fallback;
}


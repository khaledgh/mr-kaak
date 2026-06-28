import axios from "axios";
import { useAuth } from "@/stores/auth";
import type { Envelope, TokenPair } from "./types";

export const http = axios.create({ baseURL: import.meta.env.VITE_API_BASE_URL ?? "/api/v1" });

http.interceptors.request.use((config) => {
  const { accessToken } = useAuth.getState();
  if (accessToken) config.headers.Authorization = `Bearer ${accessToken}`;
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
        // Reuse the shared instance so refresh follows the same baseURL/proxy
        // path as every other call (avoids hitting the wrong origin).
        const { data } = await http.post<Envelope<{ tokens: TokenPair }>>(
          "/auth/refresh",
          { refresh_token: token },
          { _retried: true } as never,
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

export function unwrap<T>(env: Envelope<T>): T {
  if (env.error) throw new Error(env.error.message);
  return env.data as T;
}

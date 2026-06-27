import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { TokenPair, User } from "@/api/types";

// Auth store: persists tokens + user to localStorage. The access token is
// attached to API requests by the axios interceptor (see api/client.ts).
interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  setAuth: (user: User, tokens: TokenPair) => void;
  setTokens: (tokens: TokenPair) => void;
  logout: () => void;
}

export const useAuth = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      setAuth: (user, tokens) =>
        set({ user, accessToken: tokens.access_token, refreshToken: tokens.refresh_token }),
      setTokens: (tokens) =>
        set({ accessToken: tokens.access_token, refreshToken: tokens.refresh_token }),
      logout: () => set({ user: null, accessToken: null, refreshToken: null }),
    }),
    { name: "auth" },
  ),
);

import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

export type AuthUser = {
  id: string;
  email: string;
  username: string;
};

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  setSession: (session: { token: string; user: AuthUser }) => void;
  clearSession: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setSession: (session) => set({ token: session.token, user: session.user }),
      clearSession: () => set({ token: null, user: null }),
    }),
    {
      name: "auth-session",
      storage:
        typeof window !== "undefined"
          ? createJSONStorage(() => localStorage)
          : undefined,
    }
  )
);

export const authStoreApi = useAuthStore;

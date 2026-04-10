import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { ReactNode } from "react";
import type { User } from "../lib/types";
import * as authApi from "../lib/api/auth";
import { setAccessToken } from "../lib/authToken";

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, displayName: string, password: string) => Promise<void>;
  logout: () => void;
  getToken: () => string | null;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const tokenRef = useRef<string | null>(null);

  const setToken = useCallback((token: string | null) => {
    tokenRef.current = token;
    setAccessToken(token);
  }, []);

  // Attempt session restore on mount via refresh cookie.
  useEffect(() => {
    let cancelled = false;
    authApi
      .refresh()
      .then(async (res) => {
        if (cancelled) return;
        setToken(res.access_token);
        const me = await authApi.getMe(res.access_token);
        if (!cancelled) setUser(me);
      })
      .catch(() => {
        // No valid session — stay logged out.
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [setToken]);

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await authApi.login({ email, password });
      setToken(res.access_token);
      setUser(res.user);
    },
    [setToken],
  );

  const register = useCallback(
    async (email: string, displayName: string, password: string) => {
      const res = await authApi.register({
        email,
        display_name: displayName,
        password,
      });
      setToken(res.access_token);
      setUser(res.user);
    },
    [setToken],
  );

  const logout = useCallback(() => {
    setToken(null);
    setUser(null);
  }, [setToken]);

  const getToken = useCallback(() => tokenRef.current, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: !!user,
      isLoading,
      login,
      register,
      logout,
      getToken,
    }),
    [user, isLoading, login, register, logout, getToken],
  );

  return (
    <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}

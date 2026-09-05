import { createContext, useContext, useMemo, useState, useCallback, type ReactNode } from 'react';
import { api } from '../../shared/api/http';
import { TOKEN_KEY, USER_KEY } from '../../shared/config';
import type { User } from '../../shared/types';

type AuthState = {
  token: string | null;
  user: User | null;
  login: (login: string, password: string) => Promise<void>;
  register: (confirmToken: string, username: string, password: string, fullName?: string) => Promise<void>;
  sendCode: (login: string) => Promise<string | null>;
  verifyCode: (login: string, code: string) => Promise<string>;
  logout: () => void;
};

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY));
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? JSON.parse(raw) : null;
  });

  const persist = useCallback((t: string, u: User) => {
    localStorage.setItem(TOKEN_KEY, t);
    localStorage.setItem(USER_KEY, JSON.stringify(u));
    setToken(t);
    setUser(u);
  }, []);

  const login = useCallback(async (loginValue: string, password: string) => {
    const res = await api<{ token: string; user: { id: string; username: string; email?: string; phone?: string; fullName?: string } }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ login: loginValue, password }),
    });
    const profile = await api<User>('/api/users/me', {
      headers: { Authorization: `Bearer ${res.token}` },
    }).catch(() => ({
      id: String(res.user.id),
      username: res.user.username,
      displayName: res.user.fullName || res.user.username,
      email: res.user.email,
      phone: res.user.phone,
    }));
    persist(res.token, { ...profile, id: String(profile.id) });
  }, [persist]);

  const register = useCallback(async (confirmToken: string, username: string, password: string, fullName?: string) => {
    const res = await api<{ token: string; user: { id: string; username: string; email?: string; phone?: string; fullName?: string } }>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ confirmToken, username, password, fullName }),
    });
    persist(res.token, {
      id: String(res.user.id),
      username: res.user.username,
      displayName: res.user.fullName || res.user.username,
      email: res.user.email,
      phone: res.user.phone,
    });
  }, [persist]);

  const sendCode = useCallback(async (loginValue: string) => {
    const res = await api<{ debugCode?: string }>('/api/auth/send', {
      method: 'POST',
      body: JSON.stringify({ login: loginValue }),
    });
    return res.debugCode ?? null;
  }, []);

  const verifyCode = useCallback(async (loginValue: string, code: string) => {
    const res = await api<{ confirmToken: string }>('/api/auth/verify-code', {
      method: 'POST',
      body: JSON.stringify({ login: loginValue, code }),
    });
    return res.confirmToken;
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setToken(null);
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({ token, user, login, register, sendCode, verifyCode, logout }),
    [token, user, login, register, sendCode, verifyCode, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('AuthProvider missing');
  return ctx;
}

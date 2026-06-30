import { create } from 'zustand';
import type { User, LoginRequest } from '../types';
import { getToken, getUser, setToken, setUser, removeToken, removeUser } from '../utils';
import { authApi } from '../api/auth';

interface AuthState {
  token: string | null;
  user: User | null;
  loading: boolean;
  login: (data: LoginRequest) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
  hasPermission: (perm: string) => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: getToken(),
  user: getUser(),
  loading: false,

  login: async (data: LoginRequest) => {
    set({ loading: true });
    try {
      const res = await authApi.login(data);
      setToken(res.token);
      setUser(res.user);
      set({ token: res.token, user: res.user, loading: false });
    } catch {
      set({ loading: false });
      throw new Error('登录失败');
    }
  },

  logout: () => {
    removeToken();
    removeUser();
    set({ token: null, user: null });
  },

  fetchUser: async () => {
    try {
      const user = await authApi.me();
      setUser(user);
      set({ user });
    } catch {
      get().logout();
    }
  },

  hasPermission: (perm: string) => {
    const { user } = get();
    if (!user?.role) return false;
    const perms = user.role.permissions;
    if (!perms) return false;
    if (perms === '*') return true;
    return perms.split(',').includes(perm);
  },
}));

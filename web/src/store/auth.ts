import { create } from 'zustand';
import { AuthUser, login, logout as logoutRequest, me, refresh, register } from '@/api/auth';

type AuthStatus = 'idle' | 'loading' | 'authenticated' | 'unauthenticated';

interface AuthState {
  accessToken: string | null;
  user: AuthUser | null;
  status: AuthStatus;
  error: string | null;
  bootstrap: () => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshAccessToken: () => Promise<string>;
  loadUser: () => Promise<void>;
  loadUserWithToken: (accessToken: string) => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  user: null,
  status: 'idle',
  error: null,
  bootstrap: async () => {
    if (get().status !== 'idle') {
      return;
    }
    set({ status: 'loading', error: null });
    try {
      const accessToken = await get().refreshAccessToken();
      await get().loadUserWithToken(accessToken);
    } catch (error) {
      set({ accessToken: null, user: null, status: 'unauthenticated', error: null });
    }
  },
  login: async (email, password) => {
    set({ status: 'loading', error: null });
    try {
      const response = await login(email, password);
      set({ accessToken: response.accessToken, user: response.user, status: 'authenticated', error: null });
    } catch (error) {
      set({
        accessToken: null,
        user: null,
        status: 'unauthenticated',
        error: error instanceof Error ? error.message : 'Login failed',
      });
      throw error;
    }
  },
  register: async (email, password) => {
    set({ status: 'loading', error: null });
    try {
      const response = await register(email, password);
      set({ accessToken: response.accessToken, user: response.user, status: 'authenticated', error: null });
    } catch (error) {
      set({
        accessToken: null,
        user: null,
        status: 'unauthenticated',
        error: error instanceof Error ? error.message : 'Register failed',
      });
      throw error;
    }
  },
  logout: async () => {
    try {
      await logoutRequest();
    } finally {
      set({ accessToken: null, user: null, status: 'unauthenticated', error: null });
    }
  },
  refreshAccessToken: async () => {
    const response = await refresh();
    set({ accessToken: response.accessToken });
    return response.accessToken;
  },
  loadUser: async () => {
    const accessToken = get().accessToken;
    if (!accessToken) {
      throw new Error('No access token');
    }
    await get().loadUserWithToken(accessToken);
  },
  loadUserWithToken: async (accessToken: string) => {
    try {
      const response = await me(
        accessToken,
        async () => get().refreshAccessToken(),
        async () => get().logout()
      );
      set({
        accessToken: response.accessToken,
        user: response.user,
        status: 'authenticated',
        error: null,
      });
    } catch (error) {
      set({ accessToken: null, user: null, status: 'unauthenticated', error: null });
      throw error;
    }
  },
}));

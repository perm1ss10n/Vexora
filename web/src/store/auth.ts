import { create } from 'zustand';

const TOKEN_KEY = 'konyx-token';
const USER_KEY = 'konyx-user';

interface AuthState {
  token: string | null;
  email: string | null;
  login: (email: string, password: string) => void;
  register: (email: string, password: string) => void;
  logout: () => void;
}

const getInitialState = () => {
  const token = localStorage.getItem(TOKEN_KEY);
  const email = localStorage.getItem(USER_KEY);
  return { token, email };
};

export const useAuthStore = create<AuthState>((set) => ({
  ...getInitialState(),
  login: (email) => {
    const token = `token-${Date.now()}`;
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, email);
    set({ token, email });
  },
  register: (email) => {
    const token = `token-${Date.now()}`;
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, email);
    set({ token, email });
  },
  logout: () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    set({ token: null, email: null });
  },
}));

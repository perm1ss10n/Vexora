import { apiRequest, apiRequestWithAuth } from './client';

export interface AuthUser {
  id: string;
  email: string;
}

interface AuthResponse {
  user: AuthUser;
  accessToken: string;
}

interface RefreshResponse {
  accessToken: string;
}

interface MeResponse {
  user: AuthUser;
}

export const register = (email: string, password: string) =>
  apiRequest<AuthResponse>('/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });

export const login = (email: string, password: string) =>
  apiRequest<AuthResponse>('/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });

export const refresh = () => apiRequest<RefreshResponse>('/v1/auth/refresh', { method: 'POST' });

export const logout = () => apiRequest<void>('/v1/auth/logout', { method: 'POST' });

export const me = async (
  accessToken: string,
  refreshToken?: () => Promise<string>,
  onUnauthorized?: () => Promise<void>
) => {
  const { data, accessToken: nextToken } = await apiRequestWithAuth<MeResponse>(
    '/v1/auth/me',
    accessToken,
    { method: 'GET' },
    refreshToken,
    onUnauthorized
  );
  return { user: data.user, accessToken: nextToken };
};

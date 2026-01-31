export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

type RequestOptions = Omit<RequestInit, 'headers'> & { headers?: Record<string, string> };

export const apiRequest = async <T>(path: string, options: RequestOptions = {}): Promise<T> => {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include',
  });

  if (!response.ok) {
    const message = await readErrorMessage(response);
    throw new ApiError(response.status, message);
  }

  if (response.status === 204) {
    return {} as T;
  }

  return (await response.json()) as T;
};

export const apiRequestWithAuth = async <T>(
  path: string,
  accessToken: string,
  options: RequestOptions = {},
  refresh?: () => Promise<string>,
  onUnauthorized?: () => Promise<void>
): Promise<{ data: T; accessToken: string }> => {
  try {
    const data = await apiRequest<T>(path, {
      ...options,
      headers: {
        ...(options.headers ?? {}),
        Authorization: `Bearer ${accessToken}`,
      },
    });
    return { data, accessToken };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401 && refresh) {
      try {
        const newToken = await refresh();
        const data = await apiRequest<T>(path, {
          ...options,
          headers: {
            ...(options.headers ?? {}),
            Authorization: `Bearer ${newToken}`,
          },
        });
        return { data, accessToken: newToken };
      } catch (refreshError) {
        if (onUnauthorized) {
          await onUnauthorized();
        }
        throw refreshError;
      }
    }
    throw error;
  }
};

const readErrorMessage = async (response: Response): Promise<string> => {
  try {
    const data = await response.json();
    if (typeof data?.message === 'string') {
      return data.message;
    }
  } catch {
    // ignore
  }
  return response.statusText || 'Request failed';
};

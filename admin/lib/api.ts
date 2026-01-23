const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080';

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  body?: object;
  headers?: Record<string, string>;
}

export async function api<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> {
  const { method = 'GET', body, headers = {} } = options;

  const config: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      ...headers,
    },
  };

  if (body) {
    config.body = JSON.stringify(body);
  }

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);
    
    const text = await response.text();

    let data: unknown;
    try {
      data = JSON.parse(text);
    } catch {
      // Response is not valid JSON
      return {
        error: response.ok ? 'Invalid response from server' : text || `Request failed with status ${response.status}`,
      };
    }

    if (!response.ok) {
      // Extract error message from various possible response formats
      if (data && typeof data === 'object') {
        const errorObj = data as Record<string, unknown>;
        if (typeof errorObj.message === 'string') return { error: errorObj.message };
        if (typeof errorObj.error === 'string') return { error: errorObj.error };
      }
      return { error: `Request failed with status ${response.status}` };
    }

    return { data: data as T };
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Network error';
    return { error: message };
  }
}

interface RegisterPayload {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

interface AuthResponse {
  user: {
    id: string;
    email: string;
  };
  token: string;
}

export const authApi = {
  register: (payload: RegisterPayload) =>
    api<AuthResponse>('/auth/register', {
      method: 'POST',
      body: payload,
    }),

  login: (email: string, password: string) =>
    api<AuthResponse>('/auth/login', {
      method: 'POST',
      body: { email, password },
    }),
};

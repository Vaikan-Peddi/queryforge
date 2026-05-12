import type { AIHealth, DatabaseSchema, HistoryItem, QueryResult, User, Workspace } from './types';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

type TokenPair = { access_token: string; refresh_token: string; user: User };

export function getSession(): TokenPair | null {
  if (typeof localStorage === 'undefined') return null;
  const raw = localStorage.getItem('queryforge_session');
  return raw ? JSON.parse(raw) : null;
}

export function saveSession(session: TokenPair) {
  localStorage.setItem('queryforge_session', JSON.stringify(session));
}

export function clearSession() {
  localStorage.removeItem('queryforge_session');
}

async function request<T>(path: string, options: RequestInit = {}, retry = true): Promise<T> {
  const session = getSession();
  const headers = new Headers(options.headers);
  if (!(options.body instanceof FormData)) headers.set('Content-Type', 'application/json');
  if (session?.access_token) headers.set('Authorization', `Bearer ${session.access_token}`);
  const response = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (response.status === 401 && retry && session?.refresh_token) {
    const refreshed = await fetch(`${API_BASE}/api/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: session.refresh_token })
    });
    if (refreshed.ok) {
      saveSession(await refreshed.json());
      return request<T>(path, options, false);
    }
    clearSession();
  }
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(error.error || 'Request failed');
  }
  if (response.status === 204) return undefined as T;
  return response.json();
}

export const api = {
  register: (name: string, email: string, password: string) =>
    request<TokenPair>('/api/auth/register', { method: 'POST', body: JSON.stringify({ name, email, password }) }),
  login: (email: string, password: string) =>
    request<TokenPair>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
  logout: (refreshToken: string) =>
    request('/api/auth/logout', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) }),
  workspaces: () => request<{ workspaces: Workspace[] }>('/api/workspaces'),
  createWorkspace: (name: string) =>
    request<Workspace>('/api/workspaces', { method: 'POST', body: JSON.stringify({ name }) }),
  workspace: (id: string) => request<Workspace>(`/api/workspaces/${id}`),
  deleteWorkspace: (id: string) => request(`/api/workspaces/${id}`, { method: 'DELETE' }),
  uploadDatabase: (id: string, file: File) => {
    const form = new FormData();
    form.set('file', file);
    return request<Workspace>(`/api/workspaces/${id}/upload`, { method: 'POST', body: form });
  },
  schema: (id: string) => request<DatabaseSchema>(`/api/workspaces/${id}/schema`),
  generate: (id: string, question: string) =>
    request<{ sql: string; explanation: string; confidence: number }>(`/api/workspaces/${id}/query/generate`, {
      method: 'POST',
      body: JSON.stringify({ question })
    }),
  execute: (id: string, sql: string) =>
    request<QueryResult>(`/api/workspaces/${id}/query/execute`, { method: 'POST', body: JSON.stringify({ sql }) }),
  history: (id: string) => request<{ history: HistoryItem[] }>(`/api/workspaces/${id}/history?limit=30`),
  aiHealth: () => request<AIHealth>('/api/ai/health')
};

import { User } from "./types";

const BASE = process.env.NEXT_API_BASE_URL || 'http://localhost:8080';

async function req<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init.headers ?? {}) },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  signup: (data: { email: string; username: string; display_name: string; password: string }) =>
    req<User>("/api/auth/signup", { method: "POST", body: JSON.stringify(data) }),
  login: (data: { email: string; password: string }) =>
    req<User>("/api/auth/login", { method: "POST", body: JSON.stringify(data) }),
  logout: () => req<void>("/api/auth/logout", { method: "POST" }),
  me: () => req<User>("/api/auth/me"),
};
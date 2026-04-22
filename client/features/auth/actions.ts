"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import type { LoginInput, SignupInput, User } from "./types";

const API_BASE = process.env.API_BASE_URL ?? "http://localhost:8080";
const SESSION_COOKIE = "onetube_session";

export type ActionResult = { error: string } | undefined;

async function forwardSetCookies(res: Response): Promise<void> {
  const jar = await cookies();
  const setCookies = res.headers.getSetCookie();
  for (const raw of setCookies) {
    const [pair, ...attrs] = raw.split(";").map((p) => p.trim());
    const eq = pair.indexOf("=");
    if (eq < 0) continue;
    const name = pair.slice(0, eq);
    const value = pair.slice(eq + 1);

    const opts: Parameters<typeof jar.set>[2] = {};
    for (const attr of attrs) {
      const [k, v] = attr.split("=");
      const key = k.toLowerCase();
      if (key === "path") opts.path = v;
      else if (key === "domain") opts.domain = v;
      else if (key === "max-age") opts.maxAge = Number(v);
      else if (key === "expires") opts.expires = new Date(v);
      else if (key === "httponly") opts.httpOnly = true;
      else if (key === "secure") opts.secure = true;
      else if (key === "samesite") {
        const s = v?.toLowerCase();
        if (s === "lax" || s === "strict" || s === "none") opts.sameSite = s;
      }
    }
    jar.set(name, value, opts);
  }
}

async function call<T>(
  path: string,
  init: RequestInit = {},
): Promise<{ ok: boolean; status: number; res: Response; data: T | null }> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...(init.headers ?? {}) },
    cache: "no-store",
  });
  const data = res.status === 204 ? null : ((await res.json().catch(() => null)) as T | null);
  return { ok: res.ok, status: res.status, res, data };
}

export async function signupAction(_prev: ActionResult, formData: FormData): Promise<ActionResult> {
  const input: SignupInput = {
    email: String(formData.get("email") ?? "").trim(),
    username: String(formData.get("username") ?? "").trim(),
    display_name: String(formData.get("display_name") ?? "").trim(),
    password: String(formData.get("password") ?? ""),
  };
  if (!input.email || !input.username || !input.display_name || !input.password) {
    return { error: "All fields are required" };
  }
  const { ok, status, res, data } = await call<{ error?: string }>("/api/auth/signup", {
    method: "POST",
    body: JSON.stringify(input),
  });
  if (!ok) return { error: data?.error ?? `Signup failed (${status})` };
  await forwardSetCookies(res);
  redirect("/");
}

export async function loginAction(_prev: ActionResult, formData: FormData): Promise<ActionResult> {
  const input: LoginInput = {
    email: String(formData.get("email") ?? "").trim(),
    password: String(formData.get("password") ?? ""),
  };
  if (!input.email || !input.password) {
    return { error: "Email and password are required" };
  }
  const { ok, res, data } = await call<{ error?: string }>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
  if (!ok) return { error: data?.error ?? "Invalid email or password" };
  await forwardSetCookies(res);
  redirect("/");
}

export async function logoutAction(): Promise<void> {
  const jar = await cookies();
  const session = jar.get(SESSION_COOKIE);
  if (session) {
    await fetch(`${API_BASE}/api/auth/logout`, {
      method: "POST",
      headers: { Cookie: `${session.name}=${session.value}` },
      cache: "no-store",
    }).catch(() => null);
  }
  jar.delete(SESSION_COOKIE);
}

export async function getCurrentUser(): Promise<User | null> {
  const jar = await cookies();
  const session = jar.get(SESSION_COOKIE);
  if (!session) return null;
  const res = await fetch(`${API_BASE}/api/auth/me`, {
    headers: { Cookie: `${session.name}=${session.value}` },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return (await res.json()) as User;
}

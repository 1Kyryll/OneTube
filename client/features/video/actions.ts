"use server"; 

import { cookies } from "next/headers"

import type {
    CreateVideoResponse, 
    FeedPage, 
    WatchResponse, 
} from "./types"

const API_BASE = process.env.API_BASE_URL ?? "http://localhost:8080"
const SESSION_COOKIE = "onetube_session"

async function authHeader(): Promise<Record<string, string>> {
    const jar = await cookies(); 
    const session = jar.get(SESSION_COOKIE); 

    return session ? { Cookie: `${session.name}=${session.value}` } : {}; 
}

async function jsonOrThrow<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || res.statusText);
  }
  return res.json() as Promise<T>;
}


export async function createUploadIntent(input: {
    title: string; 
    description: string;
    visibility: "public" | "unlisted" | "private"; 
    filename: string; 
    contentType: string;
}): Promise<CreateVideoResponse> {
    const res = await fetch(`${API_BASE}/api/videos`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...(await authHeader()) },
        body: JSON.stringify({
        title: input.title,
        description: input.description,
        visibility: input.visibility,
        filename: input.filename,
        content_type: input.contentType,
        }),
        cache: "no-store",
    });
    
    return jsonOrThrow<CreateVideoResponse>(res);
}

export async function completeUpload(videoID: string, extension: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/videos/${videoID}/complete`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...(await authHeader()) },
    body: JSON.stringify({ extension }),
    cache: "no-store",
  });
  if (!res.ok) throw new Error(await res.text());
}

export async function getWatch(
  videoID: string,
): Promise<WatchResponse | { status: string }> {
  const res = await fetch(`${API_BASE}/api/videos/${videoID}`, {
    headers: { ...(await authHeader()) },
    cache: "no-store",
  });
  return jsonOrThrow<WatchResponse | { status: string }>(res);
}

export async function getFeed(cursor?: string): Promise<FeedPage> {
  const qs = new URLSearchParams({ limit: "24" });
  if (cursor) qs.set("cursor", cursor);
  const res = await fetch(`${API_BASE}/api/videos/feed?${qs.toString()}`, {
    cache: "no-store",
  });
  return jsonOrThrow<FeedPage>(res);
}

export async function recordView(videoID: string): Promise<void> {
  await fetch(`${API_BASE}/api/videos/${videoID}/view`, {
    method: "POST",
    cache: "no-store",
  }).catch(() => null);
}

export async function cursorToString(
  c: { published_at: string; id: string } | undefined,
): Promise<string | undefined> {
  if (!c) return undefined;
  return `${c.published_at}|${c.id}`;
}
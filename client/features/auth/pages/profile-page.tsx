"use client";

import Image from "next/image";
import { useEffect, useRef, useState } from "react";

import { completeAvatarUpload, createAvatarUploadIntent } from "../actions";
import { useAuth } from "../auth-provider";

const MAX_AVATAR_BYTES = 5 * 1024 * 1024;

export function ProfilePage() {
  const { user, loading, refresh } = useAuth();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [file, setFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!file) {
      setPreviewUrl(null);
      return;
    }
    const url = URL.createObjectURL(file);
    setPreviewUrl(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  if (loading) return <main style={pageStyle}>Loading…</main>;
  if (!user) return <main style={pageStyle}>You must be logged in.</main>;

  function onPick(e: React.ChangeEvent<HTMLInputElement>) {
    setError(null);
    const picked = e.target.files?.[0] ?? null;
    if (!picked) {
      setFile(null);
      return;
    }
    if (!picked.type.startsWith("image/")) {
      setError("File must be an image.");
      setFile(null);
      return;
    }
    if (picked.size > MAX_AVATAR_BYTES) {
      setError("Image must be under 5 MB.");
      setFile(null);
      return;
    }
    setFile(picked);
  }

  function onClear() {
    setFile(null);
    setError(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  async function onSave() {
    if (!file) return;
    setSaving(true);
    setError(null);
    try {
      const intent = await createAvatarUploadIntent(file.type);
      if ("error" in intent) throw new Error(intent.error);

      const put = await fetch(intent.url, {
        method: "PUT",
        headers: { "Content-Type": file.type },
        body: file,
      });
      if (!put.ok) throw new Error(`Upload to storage failed (${put.status})`);

      const complete = await completeAvatarUpload();
      if (complete?.error) throw new Error(complete.error);

      await refresh();
      onClear();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Upload failed");
    } finally {
      setSaving(false);
    }
  }

  const displayedAvatar = previewUrl ?? user.avatar ?? null;

  return (
    <main style={pageStyle}>
      <h1 style={{ marginBottom: 24 }}>Profile</h1>

      <section style={cardStyle}>
        <div style={avatarRowStyle}>
          {displayedAvatar ? (
            <Image
              src={displayedAvatar}
              alt="Avatar"
              width={96}
              height={96}
              unoptimized
              style={{ borderRadius: "50%", objectFit: "cover", background: "#222" }}
            />
          ) : (
            <div style={{ ...avatarStyle, ...avatarPlaceholderStyle }}>
              {user.display_name.slice(0, 1).toUpperCase()}
            </div>
          )}

          <div style={{ display: "grid", gap: 4 }}>
            <strong style={{ fontSize: 16 }}>{user.display_name}</strong>
            <span style={{ opacity: 0.7, fontSize: 14 }}>@{user.username}</span>
            <span style={{ opacity: 0.5, fontSize: 12 }}>{user.email}</span>
          </div>
        </div>

        <div style={{ display: "grid", gap: 8, marginTop: 24 }}>
          <label style={{ fontSize: 14 }}>Change avatar</label>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={onPick}
            disabled={saving}
            style={{ fontSize: 14 }}
          />
          {file && (
            <p style={{ margin: 0, fontSize: 12, opacity: 0.7 }}>
              {file.name} · {(file.size / 1024).toFixed(0)} KB
            </p>
          )}
          {error && <p style={{ margin: 0, color: "crimson", fontSize: 13 }}>{error}</p>}

          <div style={{ display: "flex", gap: 8, marginTop: 8 }}>
            <button
              type="button"
              onClick={onSave}
              disabled={!file || saving}
              style={primaryButtonStyle}
            >
              {saving ? "Saving…" : "Save"}
            </button>
            <button
              type="button"
              onClick={onClear}
              disabled={!file || saving}
              style={secondaryButtonStyle}
            >
              Cancel
            </button>
          </div>
        </div>
      </section>
    </main>
  );
}

const pageStyle: React.CSSProperties = {
  maxWidth: 520,
  margin: "64px auto",
  padding: 16,
  fontFamily: "system-ui, sans-serif",
};

const cardStyle: React.CSSProperties = {
  padding: 20,
  border: "1px solid #2a2a2a",
  borderRadius: 8,
  background: "#111",
};

const avatarRowStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: 16,
};

const avatarStyle: React.CSSProperties = {
  width: 96,
  height: 96,
  borderRadius: "50%",
  objectFit: "cover",
  background: "#222",
};

const avatarPlaceholderStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  fontSize: 36,
  fontWeight: 700,
  color: "#888",
};

const primaryButtonStyle: React.CSSProperties = {
  padding: "8px 16px",
  border: "none",
  borderRadius: 4,
  background: "#fff",
  color: "#111",
  cursor: "pointer",
  fontSize: 14,
  fontWeight: 600,
};

const secondaryButtonStyle: React.CSSProperties = {
  padding: "8px 16px",
  border: "1px solid #333",
  borderRadius: 4,
  background: "transparent",
  color: "#fff",
  cursor: "pointer",
  fontSize: 14,
};

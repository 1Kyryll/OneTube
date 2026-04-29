"use client";

import Link from "next/link";
import { useTransition } from "react";

import { useAuth } from "@/features/auth/auth-provider";

export function NavBar() {
  const { user, loading, logout } = useAuth();
  const [pending, startTransition] = useTransition();

  return (
    <header style={headerStyle}>
      <Link href="/" style={brandStyle}>
        OneTube
      </Link>

      <nav style={navStyle}>
        {!loading && user && (
          <Link href="/upload" style={linkStyle}>
            Upload
          </Link>
        )}

        {loading ? null : user ? (
          <>
            <Link href="/profile" style={profileLinkStyle}>
              {user.avatar ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={user.avatar} alt="" style={avatarStyle} />
              ) : (
                <span style={avatarPlaceholderStyle}>
                  {user.display_name.slice(0, 1).toUpperCase()}
                </span>
              )}
              <span style={{ opacity: 0.8 }}>@{user.username}</span>
            </Link>
            <button
              onClick={() => startTransition(() => logout())}
              disabled={pending}
              style={buttonStyle}
            >
              {pending ? "…" : "Log out"}
            </button>
          </>
        ) : (
          <>
            <Link href="/login" style={linkStyle}>
              Log in
            </Link>
            <Link href="/signup" style={linkStyle}>
              Sign up
            </Link>
          </>
        )}
      </nav>
    </header>
  );
}

const headerStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  padding: "12px 24px",
  borderBottom: "1px solid #222",
  background: "#0f0f0f",
  color: "#fff",
};

const brandStyle: React.CSSProperties = {
  fontWeight: 700,
  fontSize: 18,
  color: "#fff",
  textDecoration: "none",
};

const navStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: 16,
};

const linkStyle: React.CSSProperties = {
  color: "#fff",
  textDecoration: "none",
};

const buttonStyle: React.CSSProperties = {
  padding: "6px 12px",
  border: "none",
  borderRadius: 4,
  background: "#fff",
  color: "#111",
  cursor: "pointer",
  fontWeight: 600,
};

const profileLinkStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: 8,
  color: "#fff",
  textDecoration: "none",
};

const avatarStyle: React.CSSProperties = {
  width: 28,
  height: 28,
  borderRadius: "50%",
  objectFit: "cover",
  background: "#222",
};

const avatarPlaceholderStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: 28,
  height: 28,
  borderRadius: "50%",
  background: "#222",
  fontSize: 13,
  fontWeight: 700,
};

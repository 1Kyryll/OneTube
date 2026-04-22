"use client";

import Link from "next/link";
import { useTransition } from "react";

import { useAuth } from "@/features/auth/auth-provider";

export default function Home() {
  const { user, loading, logout } = useAuth();
  const [pending, startTransition] = useTransition();

  if (loading) {
    return (
      <main style={pageStyle}>
        <p>Loading…</p>
      </main>
    );
  }

  return (
    <main style={pageStyle}>
      <h1>OneTube</h1>
      {user ? (
        <>
          <p>
            Signed in as <strong>{user.display_name}</strong> (@{user.username})
          </p>
          <button
            onClick={() => startTransition(() => logout())}
            disabled={pending}
            style={buttonStyle}
          >
            {pending ? "Logging out…" : "Log out"}
          </button>
        </>
      ) : (
        <p>
          <Link href="/login">Log in</Link> · <Link href="/signup">Sign up</Link>
        </p>
      )}
    </main>
  );
}

const pageStyle: React.CSSProperties = {
  padding: 32,
  fontFamily: "system-ui, sans-serif",
};

const buttonStyle: React.CSSProperties = {
  padding: "8px 16px",
  border: "none",
  borderRadius: 4,
  background: "#111",
  color: "#fff",
  cursor: "pointer",
};

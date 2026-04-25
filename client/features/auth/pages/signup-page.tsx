"use client";

import Link from "next/link";
import { useState, useTransition } from "react";

import { signupAction, type ActionResult } from "../actions";

export function SignupPage() {
  const [error, setError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    setError(null);
    startTransition(async () => {
      const result: ActionResult = await signupAction(undefined, form);
      if (result?.error) setError(result.error);
    });
  }

  return (
    <main
      style={{
        maxWidth: 360,
        margin: "64px auto",
        padding: 16,
        fontFamily: "system-ui, sans-serif",
      }}
    >
      <h1 style={{ marginBottom: 24 }}>Sign up</h1>
      <form onSubmit={onSubmit} style={{ display: "grid", gap: 12 }}>
        <Field label="Email">
          <input name="email" type="email" required autoComplete="email" style={inputStyle} />
        </Field>
        <Field label="Username">
          <input
            name="username"
            type="text"
            required
            autoComplete="username"
            minLength={3}
            style={inputStyle}
          />
        </Field>
        <Field label="Display name">
          <input name="display_name" type="text" required style={inputStyle} />
        </Field>
        <Field label="Password">
          <input
            name="password"
            type="password"
            required
            autoComplete="new-password"
            minLength={8}
            style={inputStyle}
          />
        </Field>
        <button type="submit" disabled={pending} style={buttonStyle}>
          {pending ? "Creating account…" : "Sign up"}
        </button>
        {error && <p style={{ color: "crimson", margin: 0 }}>{error}</p>}
      </form>
      <p style={{ marginTop: 16, fontSize: 14 }}>
        Already have an account? <Link href="/login">Log in</Link>
      </p>
    </main>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label style={{ display: "grid", gap: 4 }}>
      <span style={{ fontSize: 14 }}>{label}</span>
      {children}
    </label>
  );
}

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: 8,
  border: "1px solid #2a2a2a",
  borderRadius: 4,
  fontSize: 14,
  background: "#161616",
  color: "#fff",
};

const buttonStyle: React.CSSProperties = {
  padding: 10,
  border: "none",
  borderRadius: 4,
  background: "#fff",
  color: "#111",
  cursor: "pointer",
  fontSize: 14,
  fontWeight: 600,
};

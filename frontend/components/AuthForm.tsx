"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useActionState } from "react";
import { useAuth } from "@/hooks/useAuth";

/** Mirrors the API's minimum so an obviously short password costs no round trip. */
const MIN_PASSWORD_LENGTH = 8;

const COPY = {
  login: {
    heading: "Welcome back",
    subheading: "Sign in to pick up where you left off.",
    submit: "Sign In",
    pending: "Signing in…",
    switchPrompt: "New to Reload?",
    switchLabel: "Create an account",
    switchHref: "/register",
  },
  register: {
    heading: "Create your account",
    subheading: "One account for your whole library.",
    submit: "Create Account",
    pending: "Creating account…",
    switchPrompt: "Already have an account?",
    switchLabel: "Sign in",
    switchHref: "/login",
  },
} as const;

const fieldClass =
  "w-full rounded-2xl border border-white/10 bg-surface-container-lowest px-5 py-3 text-on-surface outline-none transition-colors placeholder:text-on-surface-variant focus:border-electric-blue";

const labelClass = "text-sm font-semibold text-on-surface-variant";

export default function AuthForm({ mode }: { mode: "login" | "register" }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { signIn, signUp } = useAuth();

  const copy = COPY[mode];
  // Only same-origin paths are honoured, so a crafted ?next= cannot bounce
  // someone off the site straight after they hand over a password.
  const next = searchParams.get("next");
  const safeNext = next?.startsWith("/") && !next.startsWith("//") ? next : null;

  const [error, submit, pending] = useActionState<string | null, FormData>(
    async (_previous, formData) => {
      const email = String(formData.get("email") ?? "");
      const password = String(formData.get("password") ?? "");
      const displayName = String(formData.get("display_name") ?? "");

      try {
        if (mode === "login") {
          await signIn(email, password);
        } else {
          await signUp({ email, password, displayName });
        }
      } catch (err) {
        // The API's message is the useful one — "email is already registered"
        // beats anything generic this form could invent.
        return err instanceof Error ? err.message : "Something went wrong";
      }

      if (mode === "register") {
        // Registering does not open a session, so the new account signs in
        // next. ?next= rides along so the original destination survives.
        router.push(
          safeNext ? `/login?next=${encodeURIComponent(safeNext)}` : "/login",
        );
        return null;
      }

      router.push(safeNext ?? "/");
      // Server Components re-render so anything session-aware picks it up.
      router.refresh();
      return null;
    },
    null,
  );

  return (
    <div className="w-full max-w-md rounded-[24px] border border-white/10 bg-surface-container-low/80 p-8 backdrop-blur-xl">
      <h1 className="text-2xl font-black tracking-tight text-on-surface">
        {copy.heading}
      </h1>
      <p className="mt-2 text-sm text-on-surface-variant">{copy.subheading}</p>

      <form action={submit} className="mt-8 flex flex-col gap-5">
        {mode === "register" && (
          <div className="flex flex-col gap-2">
            <label htmlFor="display_name" className={labelClass}>
              Display name{" "}
              <span className="font-normal text-outline">(optional)</span>
            </label>
            <input
              id="display_name"
              name="display_name"
              type="text"
              maxLength={60}
              autoComplete="nickname"
              placeholder="How other players see you"
              className={fieldClass}
            />
          </div>
        )}

        <div className="flex flex-col gap-2">
          <label htmlFor="email" className={labelClass}>
            Email
          </label>
          <input
            id="email"
            name="email"
            type="email"
            required
            autoComplete="email"
            placeholder="you@example.com"
            className={fieldClass}
          />
        </div>

        <div className="flex flex-col gap-2">
          <label htmlFor="password" className={labelClass}>
            Password
          </label>
          <input
            id="password"
            name="password"
            type="password"
            required
            minLength={mode === "register" ? MIN_PASSWORD_LENGTH : undefined}
            autoComplete={
              mode === "register" ? "new-password" : "current-password"
            }
            placeholder={
              mode === "register"
                ? `At least ${MIN_PASSWORD_LENGTH} characters`
                : "Your password"
            }
            className={fieldClass}
          />
        </div>

        {error && (
          <p
            role="alert"
            className="rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
          >
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={pending}
          className="mt-1 rounded-2xl bg-electric-blue px-5 py-3 text-sm font-bold text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {pending ? copy.pending : copy.submit}
        </button>
      </form>

      <p className="mt-6 text-sm text-on-surface-variant">
        {copy.switchPrompt}{" "}
        <Link
          href={copy.switchHref}
          className="font-semibold text-neon-lime hover:underline"
        >
          {copy.switchLabel}
        </Link>
      </p>
    </div>
  );
}

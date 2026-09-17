"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useAuth } from "@/hooks/useAuth";

export default function NavAccount() {
  const { user, loading, signOut } = useAuth();
  const router = useRouter();
  const [signingOut, setSigningOut] = useState(false);

  // Hold the space until the first session check lands, rather than showing a
  // Sign In link that turns into a name a moment later.
  if (loading) {
    return <div className="h-9 w-24" aria-hidden />;
  }

  if (!user) {
    return (
      <div className="flex items-center gap-4">
        <Link
          href="/login"
          className="text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface"
        >
          Sign In
        </Link>
        <Link
          href="/register"
          className="rounded-xl bg-electric-blue px-4 py-2 text-sm font-bold text-white transition-transform active:scale-95"
        >
          Create Account
        </Link>
      </div>
    );
  }

  async function handleSignOut() {
    setSigningOut(true);
    try {
      await signOut();
      router.push("/");
      // Server Components re-render so anything session-aware picks it up.
      router.refresh();
    } finally {
      setSigningOut(false);
    }
  }

  return (
    <div className="flex items-center gap-3">
      <Link
        href="/account"
        className="flex items-center gap-3 rounded-xl border border-white/10 px-3 py-1.5 transition-colors hover:border-white/25 hover:bg-surface-container"
      >
        <span
          aria-hidden
          className="flex h-6 w-6 items-center justify-center rounded-full bg-neon-lime text-xs font-black text-background"
        >
          {user.display_name.charAt(0).toUpperCase()}
        </span>
        <span className="max-w-[12ch] truncate text-sm font-semibold text-on-surface">
          {user.display_name}
        </span>
      </Link>
      <button
        type="button"
        onClick={handleSignOut}
        disabled={signingOut}
        className="text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface disabled:cursor-not-allowed disabled:opacity-60"
      >
        {signingOut ? "Logging out…" : "Log Out"}
      </button>
    </div>
  );
}

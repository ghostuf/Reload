"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { useAuth } from "@/hooks/useAuth";

export default function NavAccount() {
  const { user, loading, signOut } = useAuth();
  const router = useRouter();
  const [signingOut, setSigningOut] = useState(false);
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);

  /**
   * An open menu listens for the two ways a user dismisses one without
   * choosing anything. Escape puts focus back on the trigger, since the
   * element the user was on is about to stop existing.
   */
  useEffect(() => {
    if (!open) return;

    function handlePointerDown(event: MouseEvent) {
      const target = event.target as Node;
      if (
        menuRef.current?.contains(target) ||
        triggerRef.current?.contains(target)
      ) {
        return;
      }
      setOpen(false);
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key !== "Escape") return;
      setOpen(false);
      triggerRef.current?.focus();
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

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
      setOpen(false);
      router.push("/");
      // Server Components re-render so anything session-aware picks it up.
      router.refresh();
    } finally {
      setSigningOut(false);
    }
  }

  return (
    <div className="relative">
      <button
        ref={triggerRef}
        type="button"
        onClick={() => setOpen((wasOpen) => !wasOpen)}
        aria-haspopup="menu"
        aria-expanded={open}
        className={`flex items-center gap-3 rounded-xl border px-3 py-1.5 transition-colors hover:border-white/25 hover:bg-surface-container ${
          open ? "border-white/25 bg-surface-container" : "border-white/10"
        }`}
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
        <svg
          aria-hidden
          viewBox="0 0 24 24"
          className={`h-4 w-4 text-on-surface-variant transition-transform ${
            open ? "rotate-180" : ""
          }`}
          fill="none"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="m6 9.5 6 6 6-6" />
        </svg>
      </button>

      {open && (
        <div
          ref={menuRef}
          role="menu"
          aria-label="Account"
          className="absolute right-0 top-full z-50 mt-2 w-56 rounded-2xl border border-white/10 bg-surface-container-low/95 p-1.5 shadow-2xl shadow-black/50 backdrop-blur-xl"
        >
          <MenuLink href="/account" onSelect={() => setOpen(false)}>
            Account
          </MenuLink>
          <MenuLink href="/library" onSelect={() => setOpen(false)}>
            Library
          </MenuLink>
          <MenuLink href="/orders" onSelect={() => setOpen(false)}>
            Order History
          </MenuLink>

          <div aria-hidden className="my-1.5 h-px bg-white/10" />

          <button
            type="button"
            role="menuitem"
            onClick={handleSignOut}
            disabled={signingOut}
            className="w-full rounded-xl px-3 py-2 text-left text-sm font-semibold text-on-surface-variant transition-colors hover:bg-surface-container hover:text-red-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {signingOut ? "Logging out…" : "Log Out"}
          </button>
        </div>
      )}
    </div>
  );
}

function MenuLink({
  href,
  onSelect,
  children,
}: {
  href: string;
  onSelect: () => void;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      role="menuitem"
      onClick={onSelect}
      className="block rounded-xl px-3 py-2 text-sm font-semibold text-on-surface-variant transition-colors hover:bg-surface-container hover:text-on-surface"
    >
      {children}
    </Link>
  );
}

"use client";

import { useRouter } from "next/navigation";
import { useTransition } from "react";

/**
 * RefreshButton re-renders the current server page, re-running its fetches.
 * It is the retry for server-rendered pages that show their own error state
 * instead of throwing to the error boundary.
 */
export default function RefreshButton({ label = "Try again" }: { label?: string }) {
  const router = useRouter();
  const [pending, startTransition] = useTransition();

  return (
    <button
      type="button"
      onClick={() => startTransition(() => router.refresh())}
      disabled={pending}
      className="mt-3 rounded-xl bg-electric-blue px-5 py-2.5 text-sm font-bold text-white transition-transform active:scale-95 disabled:opacity-60"
    >
      {pending ? "Retrying…" : label}
    </button>
  );
}

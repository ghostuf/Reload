"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/hooks/useAuth";

/**
 * RequireAuth renders its children only for a signed-in visitor, and sends
 * everyone else to the login form with ?next= set to where they were headed.
 *
 * This is a convenience redirect, not a security boundary: the markup it wraps
 * ships to the browser either way. Anything that actually needs protecting is
 * gated by the API, which reads the session cookie this client cannot forge.
 *
 * Nothing uses it yet — there are no protected pages in this iteration beyond
 * /account, which uses it in place of the redirect it used to inline.
 */
export default function RequireAuth({
  children,
  fallback,
}: {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  const { user, loading } = useAuth();
  const router = useRouter();
  // Only the path, not the query string: reading search params would force
  // every guarded page to sit behind its own Suspense boundary.
  const pathname = usePathname();

  useEffect(() => {
    if (loading || user) return;
    router.replace(`/login?next=${encodeURIComponent(pathname)}`);
  }, [loading, user, pathname, router]);

  if (loading || !user) {
    return (
      fallback ?? (
        <div className="flex flex-1 items-center justify-center px-6 py-20">
          <p className="text-sm text-on-surface-variant">
            {loading ? "Checking your session…" : "Redirecting to sign in…"}
          </p>
        </div>
      )
    );
  }

  return <>{children}</>;
}

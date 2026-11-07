"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/hooks/useAuth";

/**
 * RequireAdmin renders its children only for an admin. Signed-out visitors go
 * to the login form (and come back here after), and signed-in non-admins go to
 * the homepage.
 *
 * This is for UX only. The admin pages hold no data of their own — everything
 * they show comes from the public games list — and every admin write is
 * checked by the API, which reads the role from the database on each request.
 *
 * A Next.js proxy cannot do this check earlier: the session cookies belong to
 * the API's origin, so they never reach the Next.js server, and the role is not
 * in the token anyway.
 */
export default function RequireAdmin({ children }: { children: React.ReactNode }) {
  const { user, loading, isAdmin } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (loading || isAdmin) return;
    if (!user) {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
    } else {
      router.replace("/");
    }
  }, [loading, user, isAdmin, pathname, router]);

  if (loading || !isAdmin) {
    return (
      <div className="flex flex-1 items-center justify-center px-6 py-20">
        <p className="text-sm text-on-surface-variant">
          {loading ? "Checking your session…" : "Redirecting…"}
        </p>
      </div>
    );
  }

  return <>{children}</>;
}

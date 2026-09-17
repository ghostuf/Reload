"use client";

import RequireAuth from "@/components/RequireAuth";
import { useAuth } from "@/hooks/useAuth";

export default function AccountPage() {
  return (
    <RequireAuth>
      <AccountDetails />
    </RequireAuth>
  );
}

function AccountDetails() {
  const { user } = useAuth();

  // RequireAuth only renders this with a session, so the null case is
  // unreachable — it is here to narrow the type, not to handle a state.
  if (!user) return null;

  return (
    <main className="flex-1">
      <section className="mx-auto max-w-[1440px] px-6 pb-8 pt-16 md:px-12 md:pt-24">
        <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
          Your Account
        </p>
        <h1 className="mt-4 text-4xl font-black tracking-tight text-on-surface md:text-5xl">
          {user.display_name}
        </h1>
      </section>

      <section className="mx-auto max-w-[1440px] px-6 pb-20 md:px-12">
        <div className="max-w-xl rounded-[24px] border border-white/10 bg-surface-container-low/80 p-8 backdrop-blur-xl">
          <dl className="flex flex-col gap-5">
            <div className="flex flex-col gap-1">
              <dt className="text-sm font-semibold text-on-surface-variant">
                Email
              </dt>
              <dd className="text-on-surface">{user.email}</dd>
            </div>
            <div className="flex flex-col gap-1">
              <dt className="text-sm font-semibold text-on-surface-variant">
                Member since
              </dt>
              <dd className="text-on-surface">
                {new Date(user.created_at).toLocaleDateString(undefined, {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </dd>
            </div>
          </dl>
        </div>
      </section>
    </main>
  );
}

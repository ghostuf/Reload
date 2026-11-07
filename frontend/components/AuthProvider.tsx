"use client";

import { createContext, useCallback, useEffect, useMemo, useState } from "react";
import * as authApi from "@/lib/auth";
import type { User } from "@/types/user";

export type AuthContextValue = {
  user: User | null;
  /** True only while the first /auth/me check is in flight. */
  loading: boolean;
  /**
   * Whether to show admin UI. Only a display hint: the API checks the role on
   * every admin request, and re-reads it from the database each time.
   */
  isAdmin: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signUp: (params: authApi.RegisterParams) => Promise<void>;
  signOut: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

export default function AuthProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [user, setUser] = useState<User | null>(null);
  // Starts true so nothing renders a signed-out state at someone who is
  // signed in: until /auth/me answers, we genuinely do not know.
  const [loading, setLoading] = useState(true);

  // The session lives in HttpOnly cookies on the API's origin, which this app
  // cannot read, so the only way to know whether one exists is to ask.
  useEffect(() => {
    let active = true;

    authApi
      .getCurrentUser()
      .then((current) => {
        if (!active) return;
        setUser(current);
      })
      .catch(() => {
        // An unreachable API is not a signed-out user, but there is nothing
        // else to show until it answers.
        if (!active) return;
        setUser(null);
      })
      .finally(() => {
        if (!active) return;
        setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  const signIn = useCallback(async (email: string, password: string) => {
    const signedIn = await authApi.login(email, password);
    setUser(signedIn);
  }, []);

  // Registering does not open a session — the API returns the new user without
  // setting cookies — so this leaves auth state alone and the caller sends the
  // new account to the login form.
  const signUp = useCallback(async (params: authApi.RegisterParams) => {
    await authApi.register(params);
  }, []);

  const signOut = useCallback(async () => {
    try {
      await authApi.logout();
    } finally {
      // Whether or not the API acknowledged it, this client is done with the
      // session; its cookies are cleared either way.
      setUser(null);
    }
  }, []);

  const isAdmin = user?.role === "admin";

  const value = useMemo<AuthContextValue>(
    () => ({ user, loading, isAdmin, signIn, signUp, signOut }),
    [user, loading, isAdmin, signIn, signUp, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

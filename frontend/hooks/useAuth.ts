"use client";

import { useContext } from "react";
import { AuthContext, type AuthContextValue } from "@/components/AuthProvider";

/** useAuth reads the signed-in user and the actions that change them. */
export function useAuth(): AuthContextValue {
  const auth = useContext(AuthContext);

  if (!auth) {
    throw new Error("useAuth must be used within an AuthProvider");
  }

  return auth;
}

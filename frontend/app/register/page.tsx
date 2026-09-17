import type { Metadata } from "next";
import { Suspense } from "react";
import AuthForm from "@/components/AuthForm";

export const metadata: Metadata = {
  title: "Create Account — Reload",
  description: "Create a Reload account to start building your library.",
};

export default function RegisterPage() {
  return (
    <main className="flex flex-1 items-center justify-center px-6 py-20 md:px-12">
      {/* AuthForm reads ?next=, so it needs a boundary to render statically. */}
      <Suspense fallback={null}>
        <AuthForm mode="register" />
      </Suspense>
    </main>
  );
}

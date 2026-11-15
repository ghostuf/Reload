import type { Metadata } from "next";
import { Suspense } from "react";
import EsewaSuccessView from "@/components/EsewaSuccessView";

export const metadata: Metadata = {
  title: "Verifying Payment — Reload",
  description: "Confirming your eSewa payment.",
};

export default function EsewaSuccessPage() {
  return (
    <main className="mx-auto flex w-full max-w-xl flex-1 items-center justify-center px-6 py-20 md:px-12">
      {/* EsewaSuccessView reads ?data=, so it needs a boundary to render statically. */}
      <Suspense fallback={null}>
        <EsewaSuccessView />
      </Suspense>
    </main>
  );
}

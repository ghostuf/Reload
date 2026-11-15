import type { Metadata } from "next";
import { Suspense } from "react";
import EsewaFailureView from "@/components/EsewaFailureView";

export const metadata: Metadata = {
  title: "Payment Cancelled — Reload",
  description: "Your eSewa payment was not completed.",
};

export default function EsewaFailurePage() {
  return (
    <main className="mx-auto flex w-full max-w-xl flex-1 items-center justify-center px-6 py-20 md:px-12">
      {/* EsewaFailureView reads ?transaction_uuid=, so it needs a boundary to render statically. */}
      <Suspense fallback={null}>
        <EsewaFailureView />
      </Suspense>
    </main>
  );
}

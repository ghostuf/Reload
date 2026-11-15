"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import RequireAuth from "@/components/RequireAuth";
import { failEsewaPayment } from "@/lib/esewa";

export default function EsewaFailureView() {
  return (
    <RequireAuth>
      <CancelledPayment />
    </RequireAuth>
  );
}

/**
 * CancelledPayment reports the abandoned or failed attempt back to the API
 * so the pending order and payment are marked cancelled rather than left to
 * expire on their own. The cart is left exactly as it was -- nothing here
 * clears it -- so the user can go straight back and try again.
 */
function CancelledPayment() {
  const searchParams = useSearchParams();
  const [reported, setReported] = useState(false);
  const started = useRef(false);

  const transactionUuid = searchParams.get("transaction_uuid");

  useEffect(() => {
    if (started.current || !transactionUuid) return;
    started.current = true;

    failEsewaPayment(transactionUuid)
      .catch(() => {
        // Nothing the user can act on here -- the order is still pending
        // and will simply expire on its own after about 30 minutes.
      })
      .finally(() => setReported(true));
  }, [transactionUuid]);

  return (
    <div className="flex flex-col items-center gap-3 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-16 text-center">
      <span
        aria-hidden
        className="flex h-12 w-12 items-center justify-center rounded-full bg-white/5 text-xl font-black text-on-surface-variant"
      >
        ×
      </span>
      <p className="text-lg font-semibold text-on-surface">
        Payment cancelled
      </p>
      <p className="max-w-md text-sm text-on-surface-variant">
        {reported
          ? "Your payment was not completed. Your cart is untouched, so you can try again whenever you're ready."
          : "Recording the cancelled payment…"}
      </p>
      <Link
        href="/cart"
        className="mt-3 rounded-xl bg-electric-blue px-5 py-2.5 text-sm font-bold text-white transition-transform active:scale-95"
      >
        Back to Cart
      </Link>
    </div>
  );
}

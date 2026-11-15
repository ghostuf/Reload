"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { Spinner } from "@/components/GameActions";
import RequireAuth from "@/components/RequireAuth";
import { useCart } from "@/hooks/useCart";
import { useLibrary } from "@/hooks/useLibrary";
import { verifyEsewaPayment } from "@/lib/esewa";

export default function EsewaSuccessView() {
  return (
    <RequireAuth>
      <VerifyingPayment />
    </RequireAuth>
  );
}

/**
 * VerifyingPayment calls POST /checkout/esewa/verify exactly once with the
 * ?data= eSewa appended to this URL, then redirects to the receipt. The
 * ref guards against React's effects running twice in development, and
 * against the user's own refresh re-triggering a call the API's own
 * idempotency would have no trouble with anyway -- belt and suspenders.
 */
function VerifyingPayment() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const cart = useCart();
  const library = useLibrary();
  const [error, setError] = useState<string | null>(null);
  const started = useRef(false);

  const data = searchParams.get("data");

  useEffect(() => {
    if (started.current || !data) return;
    started.current = true;

    verifyEsewaPayment(data)
      .then((order) => {
        // The API cleared the cart and granted the library as part of
        // completing the order; re-reading both here is what makes the
        // navbar badge and "Owned" states catch up without a reload.
        cart.clearCart();
        library.refresh().catch(() => {});
        router.replace(`/orders/${order.id}`);
      })
      .catch((err: unknown) => {
        setError(
          err instanceof Error && err.message
            ? err.message
            : "Could not verify your payment.",
        );
      });
  }, [data, cart, library, router]);

  // No ?data= at all is known the moment this renders -- derived here
  // rather than set from the effect, which only ever reports the async
  // verify call's own outcome.
  const displayError =
    error ?? (!data ? "eSewa did not send back any payment data to verify." : null);

  if (displayError) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-game-card border border-red-400/30 bg-red-400/5 px-6 py-16 text-center">
        <span
          aria-hidden
          className="flex h-12 w-12 items-center justify-center rounded-full bg-red-400/10 text-xl font-black text-red-400"
        >
          !
        </span>
        <p className="text-lg font-semibold text-on-surface">
          Payment verification failed
        </p>
        <p className="max-w-md text-sm text-on-surface-variant">{displayError}</p>
        <Link
          href="/cart"
          className="mt-3 rounded-xl bg-electric-blue px-5 py-2.5 text-sm font-bold text-white transition-transform active:scale-95"
        >
          Back to Cart
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-3 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-16 text-center">
      <Spinner />
      <p className="text-lg font-semibold text-on-surface">
        Verifying your payment…
      </p>
      <p className="text-sm text-on-surface-variant">
        Confirming with eSewa — this only takes a moment.
      </p>
    </div>
  );
}

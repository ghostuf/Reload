"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import { useAuth } from "@/hooks/useAuth";
import { useCart } from "@/hooks/useCart";
import { useLibrary } from "@/hooks/useLibrary";
import { useWishlist } from "@/hooks/useWishlist";

type Variant = "card" | "detail";

/**
 * useSignInRedirect sends a signed-out visitor to the login form, with ?next=
 * set so they land back on the game they were looking at.
 */
function useSignInRedirect() {
  const router = useRouter();
  // Only the path, matching RequireAuth: reading search params would force
  // every page that shows these buttons behind its own Suspense boundary.
  const pathname = usePathname();

  return () => router.push(`/login?next=${encodeURIComponent(pathname)}`);
}

export function Spinner() {
  return (
    <svg
      aria-hidden
      viewBox="0 0 24 24"
      className="h-4 w-4 animate-spin"
      fill="none"
    >
      <circle
        cx="12"
        cy="12"
        r="9"
        stroke="currentColor"
        strokeWidth="3"
        className="opacity-25"
      />
      <path
        d="M21 12a9 9 0 0 0-9-9"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function CartIcon({ className = "h-4 w-4" }: { className?: string }) {
  return (
    <svg
      aria-hidden
      viewBox="0 0 24 24"
      className={className}
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="9" cy="20" r="1.5" />
      <circle cx="18" cy="20" r="1.5" />
      <path d="M2.5 3h2.6l2.4 11.2a1.5 1.5 0 0 0 1.5 1.2h8.6a1.5 1.5 0 0 0 1.5-1.1L21 7H6" />
    </svg>
  );
}

export function HeartIcon({
  filled = false,
  className = "h-4 w-4",
}: {
  filled?: boolean;
  className?: string;
}) {
  return (
    <svg
      aria-hidden
      viewBox="0 0 24 24"
      className={className}
      fill={filled ? "currentColor" : "none"}
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M12 20.5s-7.5-4.6-9.3-9.4C1.5 7.8 3.6 4.5 7 4.5c2 0 3.6 1.1 5 3 1.4-1.9 3-3 5-3 3.4 0 5.5 3.3 4.3 6.6-1.8 4.8-9.3 9.4-9.3 9.4Z" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg
      aria-hidden
      viewBox="0 0 24 24"
      className="h-4 w-4"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="m5 12.5 4.5 4.5L19 7.5" />
    </svg>
  );
}

/** A cart button also appears in wishlist rows, where it sits beside Remove. */
type CartVariant = Variant | "row";

const cartButtonSize: Record<CartVariant, string> = {
  card: "w-full px-3 py-2 text-sm",
  detail: "px-6 py-3 text-base",
  row: "px-4 py-2 text-sm",
};

/**
 * CartButton adds a game to the cart. Once the game is in the cart it becomes a
 * link to /cart, since adding it a second time is not something the cart allows.
 *
 * A game already in the user's library replaces it entirely: there is nothing
 * to sell someone who owns it, and checkout would skip it anyway. Owning beats
 * being in the cart, which is the state a game bought in another tab lands in.
 */
export function CartButton({
  gameId,
  variant = "card",
}: {
  gameId: number;
  variant?: CartVariant;
}) {
  const { user } = useAuth();
  const cart = useCart();
  const library = useLibrary();
  const redirectToSignIn = useSignInRedirect();
  const [error, setError] = useState<string | null>(null);

  const pending = cart.isPending(gameId);
  const size = cartButtonSize[variant];

  if (user && library.has(gameId)) {
    return (
      <Link
        href="/library"
        className={`flex items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/5 font-bold text-on-surface transition-colors hover:border-white/25 hover:bg-surface-container ${size}`}
      >
        <span className="text-neon-lime">
          <CheckIcon />
        </span>
        Owned
      </Link>
    );
  }

  if (user && cart.has(gameId)) {
    return (
      <Link
        href="/cart"
        className={`flex items-center justify-center gap-2 rounded-xl border border-neon-lime/40 bg-neon-lime/10 font-bold text-neon-lime transition-colors hover:bg-neon-lime/20 ${size}`}
      >
        <CheckIcon />
        In Cart
      </Link>
    );
  }

  async function handleClick() {
    if (!user) {
      redirectToSignIn();
      return;
    }

    setError(null);
    try {
      await cart.addToCart(gameId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not add to cart");
    }
  }

  return (
    <div className="flex flex-col gap-1.5">
      <button
        type="button"
        onClick={handleClick}
        // While the lists load we cannot tell whether it is already in the cart
        // or already owned, and offering to sell an owned game is worse than
        // making the button wait a moment.
        disabled={cart.loading || library.loading || pending}
        title={error ?? undefined}
        className={`flex items-center justify-center gap-2 rounded-xl bg-electric-blue font-bold text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-60 disabled:active:scale-100 ${size}`}
      >
        {pending ? <Spinner /> : <CartIcon />}
        {pending ? "Adding…" : error ? "Try Again" : "Add to Cart"}
      </button>
      {error && variant === "detail" && (
        <p role="alert" className="text-sm text-red-400">
          {error}
        </p>
      )}
    </div>
  );
}

/**
 * RemoveButton takes a game off the cart or wishlist. The row disappears once
 * it succeeds, so the only states it has to show are pending and failed.
 */
export function RemoveButton({
  pending,
  onRemove,
  label,
}: {
  pending: boolean;
  onRemove: () => Promise<void>;
  /** Accessible name, e.g. "Remove Hades from cart". */
  label: string;
}) {
  const [error, setError] = useState<string | null>(null);

  async function handleClick() {
    setError(null);
    try {
      await onRemove();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not remove it");
    }
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={pending}
      aria-label={label}
      title={error ?? undefined}
      className={`flex items-center gap-2 rounded-xl border px-4 py-2 text-sm font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-60 ${
        error
          ? "border-red-400/60 text-red-400"
          : "border-white/10 text-on-surface-variant hover:border-red-400/50 hover:text-red-400"
      }`}
    >
      {pending && <Spinner />}
      {pending ? "Removing…" : error ? "Try Again" : "Remove"}
    </button>
  );
}

/**
 * WishlistButton toggles a game on the wishlist: an outlined heart adds it and
 * a filled one removes it.
 */
export function WishlistButton({
  gameId,
  variant = "card",
}: {
  gameId: number;
  variant?: Variant;
}) {
  const { user } = useAuth();
  const wishlist = useWishlist();
  const redirectToSignIn = useSignInRedirect();
  const [error, setError] = useState<string | null>(null);

  const saved = Boolean(user) && wishlist.has(gameId);
  const pending = wishlist.isPending(gameId);
  const label = saved ? "Remove from wishlist" : "Add to wishlist";

  async function handleClick() {
    if (!user) {
      redirectToSignIn();
      return;
    }

    setError(null);
    try {
      if (saved) {
        await wishlist.removeFromWishlist(gameId);
      } else {
        await wishlist.addToWishlist(gameId);
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not update your wishlist",
      );
    }
  }

  const disabled = wishlist.loading || pending;
  const icon = pending ? (
    <Spinner />
  ) : (
    <HeartIcon
      filled={saved}
      className={variant === "card" ? "h-5 w-5" : "h-4 w-4"}
    />
  );

  if (variant === "card") {
    return (
      <button
        type="button"
        onClick={handleClick}
        disabled={disabled}
        aria-label={label}
        aria-pressed={saved}
        title={error ?? label}
        className={`flex h-10 w-10 items-center justify-center rounded-full border backdrop-blur-md transition-transform active:scale-90 disabled:cursor-not-allowed disabled:opacity-60 disabled:active:scale-100 ${
          saved
            ? "border-neon-lime/40 bg-black/60 text-neon-lime"
            : error
              ? "border-red-400/60 bg-black/60 text-red-400"
              : "border-white/15 bg-black/50 text-white hover:text-neon-lime"
        }`}
      >
        {icon}
      </button>
    );
  }

  return (
    <div className="flex flex-col gap-1.5">
      <button
        type="button"
        onClick={handleClick}
        disabled={disabled}
        aria-pressed={saved}
        className={`flex items-center justify-center gap-2 rounded-xl border px-6 py-3 text-base font-bold transition-colors disabled:cursor-not-allowed disabled:opacity-60 ${
          saved
            ? "border-neon-lime/40 bg-neon-lime/10 text-neon-lime hover:bg-neon-lime/20"
            : "border-white/15 text-on-surface hover:border-white/30 hover:bg-surface-container"
        }`}
      >
        {icon}
        {saved ? "Wishlisted" : "Add to Wishlist"}
      </button>
      {error && (
        <p role="alert" className="text-sm text-red-400">
          {error}
        </p>
      )}
    </div>
  );
}

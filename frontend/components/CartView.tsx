"use client";

import Link from "next/link";
import { CartIcon, RemoveButton } from "@/components/GameActions";
import GameListItem, { GameListSkeleton } from "@/components/GameListItem";
import RequireAuth from "@/components/RequireAuth";
import { useCart } from "@/hooks/useCart";

export default function CartView() {
  return (
    <RequireAuth>
      <CartContents />
    </RequireAuth>
  );
}

function CartContents() {
  const cart = useCart();
  const count = cart.items.length;

  return (
    <main className="flex-1">
      <section className="mx-auto max-w-[1440px] px-6 pb-8 pt-16 md:px-12 md:pt-24">
        <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
          Your Cart
        </p>
        <h1 className="mt-4 text-4xl font-black tracking-tight text-on-surface md:text-5xl">
          Cart
        </h1>
        {!cart.loading && count > 0 && (
          <p className="mt-2 text-sm text-on-surface-variant">
            {count} {count === 1 ? "game" : "games"}
          </p>
        )}
      </section>

      <section className="mx-auto max-w-[1440px] px-6 pb-20 md:px-12">
        {cart.loading ? (
          <GameListSkeleton />
        ) : cart.error ? (
          <p
            role="alert"
            className="rounded-game-card border border-red-400/30 bg-red-400/5 px-6 py-5 text-sm text-red-400"
          >
            {cart.error}
          </p>
        ) : count === 0 ? (
          <EmptyCart />
        ) : (
          <div className="grid items-start gap-8 lg:grid-cols-[1fr_360px]">
            <ul className="flex flex-col gap-3">
              {cart.items.map((item) => (
                <GameListItem
                  key={item.id}
                  game={item.game}
                  actions={
                    <RemoveButton
                      label={`Remove ${item.game.title} from cart`}
                      pending={cart.isPending(item.game_id)}
                      onRemove={() => cart.removeFromCart(item.game_id)}
                    />
                  }
                />
              ))}
            </ul>

            <aside className="rounded-[24px] border border-white/10 bg-surface-container-low/80 p-6 backdrop-blur-xl lg:sticky lg:top-24">
              <h2 className="text-lg font-bold text-on-surface">Summary</h2>
              <dl className="mt-5 flex flex-col gap-3 text-sm">
                <div className="flex justify-between">
                  <dt className="text-on-surface-variant">
                    {count} {count === 1 ? "item" : "items"}
                  </dt>
                  <dd className="text-on-surface">${cart.total.toFixed(2)}</dd>
                </div>
                <div className="flex items-baseline justify-between border-t border-white/10 pt-4">
                  <dt className="font-semibold text-on-surface">Total</dt>
                  <dd className="text-2xl font-black text-on-surface">
                    ${cart.total.toFixed(2)}
                  </dd>
                </div>
              </dl>
              <Link
                href="/"
                className="mt-6 block text-center text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface"
              >
                Continue Browsing
              </Link>
            </aside>
          </div>
        )}
      </section>
    </main>
  );
}

function EmptyCart() {
  return (
    <div className="flex flex-col items-center gap-3 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-24 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-full bg-white/5 text-on-surface-variant">
        <CartIcon className="h-6 w-6" />
      </span>
      <p className="text-lg font-semibold text-on-surface">
        Your cart is empty
      </p>
      <p className="text-sm text-on-surface-variant">
        Games you add to your cart will show up here.
      </p>
      <Link
        href="/"
        className="mt-3 rounded-xl bg-electric-blue px-5 py-2.5 text-sm font-bold text-white transition-transform active:scale-95"
      >
        Browse Games
      </Link>
    </div>
  );
}

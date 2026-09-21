"use client";

import Link from "next/link";
import { CartButton, HeartIcon, RemoveButton } from "@/components/GameActions";
import GameListItem, { GameListSkeleton } from "@/components/GameListItem";
import RequireAuth from "@/components/RequireAuth";
import { useWishlist } from "@/hooks/useWishlist";

export default function WishlistView() {
  return (
    <RequireAuth>
      <WishlistContents />
    </RequireAuth>
  );
}

function WishlistContents() {
  const wishlist = useWishlist();
  const count = wishlist.items.length;

  return (
    <main className="flex-1">
      <section className="mx-auto max-w-[1440px] px-6 pb-8 pt-16 md:px-12 md:pt-24">
        <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
          Your Wishlist
        </p>
        <h1 className="mt-4 text-4xl font-black tracking-tight text-on-surface md:text-5xl">
          Wishlist
        </h1>
        {!wishlist.loading && count > 0 && (
          <p className="mt-2 text-sm text-on-surface-variant">
            {count} {count === 1 ? "game" : "games"}
          </p>
        )}
      </section>

      <section className="mx-auto max-w-[1440px] px-6 pb-20 md:px-12">
        <div className="max-w-4xl">
          {wishlist.loading ? (
            <GameListSkeleton />
          ) : wishlist.error ? (
            <p
              role="alert"
              className="rounded-game-card border border-red-400/30 bg-red-400/5 px-6 py-5 text-sm text-red-400"
            >
              {wishlist.error}
            </p>
          ) : count === 0 ? (
            <EmptyWishlist />
          ) : (
            <ul className="flex flex-col gap-3">
              {wishlist.items.map((item) => (
                <GameListItem
                  key={item.id}
                  game={item.game}
                  actions={
                    <>
                      {/* Adding to the cart leaves it on the wishlist: saving
                          a game and buying it are separate decisions. */}
                      <CartButton gameId={item.game_id} variant="row" />
                      <RemoveButton
                        label={`Remove ${item.game.title} from wishlist`}
                        pending={wishlist.isPending(item.game_id)}
                        onRemove={() =>
                          wishlist.removeFromWishlist(item.game_id)
                        }
                      />
                    </>
                  }
                />
              ))}
            </ul>
          )}
        </div>
      </section>
    </main>
  );
}

function EmptyWishlist() {
  return (
    <div className="flex flex-col items-center gap-3 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-24 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-full bg-white/5 text-on-surface-variant">
        <HeartIcon className="h-6 w-6" />
      </span>
      <p className="text-lg font-semibold text-on-surface">
        Your wishlist is empty
      </p>
      <p className="text-sm text-on-surface-variant">
        Tap the heart on any game to save it for later.
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

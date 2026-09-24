"use client";

import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { ApiError } from "@/lib/api";
import * as cartApi from "@/lib/cart";
import { getLibrary } from "@/lib/orders";
import { useAuth } from "@/hooks/useAuth";
import type { CartItem, WishlistItem } from "@/types/cart";
import type { LibraryItem } from "@/types/order";

type Collection<T> = {
  items: T[];
  /** True until the list for the current user has loaded (or failed to). */
  loading: boolean;
  /** Set when the list could not be loaded; the items are then empty. */
  error: string | null;
  has: (gameId: number) => boolean;
  /** True while an add or remove for this game is in flight. */
  isPending: (gameId: number) => boolean;
};

export type CartContextValue = Collection<CartItem> & {
  /** What the cart costs, priced the way the API prices it. */
  total: number;
  addToCart: (gameId: number) => Promise<void>;
  removeFromCart: (gameId: number) => Promise<void>;
  /**
   * Empties the cart locally, for after a checkout the API has already
   * emptied it server-side. It sends no request of its own.
   */
  clearCart: () => void;
};

export type WishlistContextValue = Collection<WishlistItem> & {
  addToWishlist: (gameId: number) => Promise<void>;
  removeFromWishlist: (gameId: number) => Promise<void>;
};

/**
 * The library is read-only: games arrive in it by being bought, never by being
 * added or taken out, so it has no mutations and no per-game pending state.
 */
export type LibraryContextValue = Omit<
  Collection<LibraryItem>,
  "isPending"
> & {
  /** Re-reads the library, after a checkout added to it. */
  refresh: () => Promise<void>;
};

export const CartContext = createContext<CartContextValue | null>(null);
export const WishlistContext = createContext<WishlistContextValue | null>(null);
export const LibraryContext = createContext<LibraryContextValue | null>(null);

type CollectionApi<T> = {
  list: () => Promise<T[]>;
  add: (gameId: number) => Promise<T>;
  remove: (gameId: number) => Promise<void>;
};

// Module-level so their identity is stable across renders.
const cartCollectionApi: CollectionApi<CartItem> = {
  list: async () => (await cartApi.getCart()).data,
  add: cartApi.addToCart,
  remove: cartApi.removeFromCart,
};

const wishlistCollectionApi: CollectionApi<WishlistItem> = {
  list: cartApi.getWishlist,
  add: cartApi.addToWishlist,
  remove: cartApi.removeFromWishlist,
};

/** One shared empty list, so "no items" keeps the same identity across renders. */
const NO_ITEMS: never[] = [];

type Loaded<T> = {
  /** The user these items belong to. */
  owner: number;
  items: T[];
  error: string | null;
};

function messageOf(err: unknown, fallback: string): string {
  return err instanceof Error && err.message ? err.message : fallback;
}

/**
 * useUserList keeps one user's list of games — cart, wishlist, or library —
 * in sync with the API.
 *
 * The loaded list is tagged with the user it belongs to rather than cleared
 * when the user changes. Anything tagged with someone else — including a
 * response that lands after sign-out — simply reads as empty, so logging out
 * clears the list without a reset step that could race an in-flight request.
 *
 * `list` must be stable across renders, since a new identity re-runs the read.
 */
function useUserList<T extends { game_id: number }>(
  userId: number | null,
  list: () => Promise<T[]>,
) {
  const [loaded, setLoaded] = useState<Loaded<T> | null>(null);

  const current = loaded !== null && loaded.owner === userId ? loaded : null;

  useEffect(() => {
    if (userId === null) return;
    let active = true;

    list()
      .then((items) => {
        if (!active) return;
        setLoaded({ owner: userId, items, error: null });
      })
      .catch((err: unknown) => {
        if (!active) return;
        setLoaded({
          owner: userId,
          items: [],
          error: messageOf(err, "Could not load your items"),
        });
      });

    return () => {
      active = false;
    };
  }, [userId, list]);

  /** update applies a change only if the list still belongs to this user. */
  const update = useCallback(
    (change: (items: T[]) => T[]) => {
      setLoaded((prev) =>
        prev && prev.owner === userId
          ? { ...prev, items: change(prev.items) }
          : prev,
      );
    },
    [userId],
  );

  /** resync replaces the list with the server's, after it proved us stale. */
  const resync = useCallback(async () => {
    const items = await list();
    update(() => items);
  }, [list, update]);

  const items: T[] = current?.items ?? NO_ITEMS;

  const has = useCallback(
    (gameId: number) => items.some((i) => i.game_id === gameId),
    [items],
  );

  return {
    items,
    // Until the user is known there is no list to wait for, so only a signed-in
    // user with nothing loaded yet counts as loading.
    loading: userId !== null && current === null,
    error: current?.error ?? null,
    has,
    update,
    resync,
  };
}

/**
 * useGameCollection adds the cart's and wishlist's mutations to a user list:
 * adding, removing, and the pending state each one is in.
 */
function useGameCollection<T extends { game_id: number }>(
  userId: number | null,
  api: CollectionApi<T>,
) {
  const { items, loading, error, has, update, resync } = useUserList(
    userId,
    api.list,
  );
  const [pending, setPending] = useState<ReadonlySet<number>>(() => new Set());

  const track = useCallback(
    async (gameId: number, request: () => Promise<void>) => {
      if (userId === null) {
        throw new Error("Sign in to manage your cart and wishlist");
      }

      setPending((prev) => new Set(prev).add(gameId));
      try {
        await request();
      } finally {
        setPending((prev) => {
          const next = new Set(prev);
          next.delete(gameId);
          return next;
        });
      }
    },
    [userId],
  );

  const add = useCallback(
    (gameId: number) =>
      track(gameId, async () => {
        try {
          const item = await api.add(gameId);
          update((items) =>
            items.some((i) => i.game_id === gameId) ? items : [...items, item],
          );
        } catch (err) {
          // Already there — added from another tab, or a double click. The
          // outcome is what the user asked for, so pick up the server's list.
          if (err instanceof ApiError && err.status === 409) {
            await resync();
            return;
          }
          throw err;
        }
      }),
    [api, track, update, resync],
  );

  const remove = useCallback(
    (gameId: number) =>
      track(gameId, async () => {
        try {
          await api.remove(gameId);
        } catch (err) {
          // Already gone, which is where the user wanted it.
          if (!(err instanceof ApiError && err.status === 404)) throw err;
        }
        update((items) => items.filter((i) => i.game_id !== gameId));
      }),
    [api, track, update],
  );

  /** clear drops the list locally, for when the server has already emptied it. */
  const clear = useCallback(() => update(() => NO_ITEMS), [update]);

  const isPending = useCallback(
    (gameId: number) => pending.has(gameId),
    [pending],
  );

  return useMemo(
    () => ({ items, loading, error, has, isPending, add, remove, clear }),
    [items, loading, error, has, isPending, add, remove, clear],
  );
}

export default function ShopProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, loading: authLoading } = useAuth();
  const userId = user?.id ?? null;

  const cart = useGameCollection(userId, cartCollectionApi);
  const wishlist = useGameCollection(userId, wishlistCollectionApi);
  const library = useUserList(userId, getLibrary);

  const total = useMemo(
    () =>
      cart.items.reduce(
        (sum, item) => sum + cartApi.effectivePrice(item.game),
        0,
      ),
    [cart.items],
  );

  // Until the session check lands we cannot know whose lists to show, so
  // report loading rather than an empty cart that fills in a moment later.
  const cartValue = useMemo<CartContextValue>(
    () => ({
      items: cart.items,
      loading: authLoading || cart.loading,
      error: cart.error,
      has: cart.has,
      isPending: cart.isPending,
      total,
      addToCart: cart.add,
      removeFromCart: cart.remove,
      clearCart: cart.clear,
    }),
    [cart, authLoading, total],
  );

  const wishlistValue = useMemo<WishlistContextValue>(
    () => ({
      items: wishlist.items,
      loading: authLoading || wishlist.loading,
      error: wishlist.error,
      has: wishlist.has,
      isPending: wishlist.isPending,
      addToWishlist: wishlist.add,
      removeFromWishlist: wishlist.remove,
    }),
    [wishlist, authLoading],
  );

  const libraryValue = useMemo<LibraryContextValue>(
    () => ({
      items: library.items,
      loading: authLoading || library.loading,
      error: library.error,
      has: library.has,
      refresh: library.resync,
    }),
    [library, authLoading],
  );

  return (
    <CartContext.Provider value={cartValue}>
      <WishlistContext.Provider value={wishlistValue}>
        <LibraryContext.Provider value={libraryValue}>
          {children}
        </LibraryContext.Provider>
      </WishlistContext.Provider>
    </CartContext.Provider>
  );
}

import { authedFetch, readError } from "@/lib/auth";
import type {
  CartItem,
  CartItemResponse,
  CartResponse,
  WishlistItem,
  WishlistItemResponse,
  WishlistResponse,
} from "@/types/cart";

/**
 * Cart and wishlist both live behind the session cookies, so every call goes
 * through authedFetch and must run in the browser for the same reason the
 * auth calls do. Failures surface as ApiError with the API's status, so callers
 * can tell a duplicate add (409) or an already-removed item (404) apart from a
 * real failure.
 */

function addRequest(gameId: number): RequestInit {
  return {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ game_id: gameId }),
  };
}

export async function getCart(): Promise<CartResponse> {
  const res = await authedFetch("/cart");
  if (!res.ok) {
    throw await readError(res, "Could not load your cart");
  }
  return res.json();
}

export async function addToCart(gameId: number): Promise<CartItem> {
  const res = await authedFetch("/cart", addRequest(gameId));
  if (!res.ok) {
    throw await readError(res, "Could not add this game to your cart");
  }
  const body: CartItemResponse = await res.json();
  return body.data;
}

export async function removeFromCart(gameId: number): Promise<void> {
  const res = await authedFetch(`/cart/${gameId}`, { method: "DELETE" });
  if (!res.ok) {
    throw await readError(res, "Could not remove this game from your cart");
  }
}

export async function getWishlist(): Promise<WishlistItem[]> {
  const res = await authedFetch("/wishlist");
  if (!res.ok) {
    throw await readError(res, "Could not load your wishlist");
  }
  const body: WishlistResponse = await res.json();
  return body.data;
}

export async function addToWishlist(gameId: number): Promise<WishlistItem> {
  const res = await authedFetch("/wishlist", addRequest(gameId));
  if (!res.ok) {
    throw await readError(res, "Could not add this game to your wishlist");
  }
  const body: WishlistItemResponse = await res.json();
  return body.data;
}

export async function removeFromWishlist(gameId: number): Promise<void> {
  const res = await authedFetch(`/wishlist/${gameId}`, { method: "DELETE" });
  if (!res.ok) {
    throw await readError(res, "Could not remove this game from your wishlist");
  }
}

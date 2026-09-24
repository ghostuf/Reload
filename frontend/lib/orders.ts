import { authedFetch, readError } from "@/lib/auth";
import type {
  LibraryItem,
  LibraryResponse,
  Order,
  OrderResponse,
  OrderSummary,
  OrdersResponse,
} from "@/types/order";

/**
 * Checkout, orders, and the library all sit behind the session cookies, so
 * every call here goes through authedFetch for the same reason the cart calls
 * do. Failures surface as ApiError carrying the API's own message and status,
 * so callers can show what actually went wrong — an empty cart and a server
 * fault should not read the same to the user.
 */

/**
 * checkout turns the cart into an order and returns it with its items and
 * their games preloaded, ready to render as a receipt.
 *
 * A 400 means the cart had nothing to charge for: it was empty, or every game
 * in it was already in the library. The message on the error says which.
 */
export async function checkout(): Promise<Order> {
  const res = await authedFetch("/checkout", { method: "POST" });
  if (!res.ok) {
    throw await readError(res, "Could not complete your purchase");
  }
  const body: OrderResponse = await res.json();
  return body.data;
}

/** getOrders lists past orders, newest first. */
export async function getOrders(): Promise<OrderSummary[]> {
  const res = await authedFetch("/orders");
  if (!res.ok) {
    throw await readError(res, "Could not load your orders");
  }
  const body: OrdersResponse = await res.json();
  return body.data;
}

/**
 * getOrder returns one order with its line items, or null when there is no
 * such order on this account. The API answers 404 both for an id that does not
 * exist and for one belonging to someone else, so there is nothing to tell
 * apart here either.
 */
export async function getOrder(id: string | number): Promise<Order | null> {
  const res = await authedFetch(`/orders/${id}`);
  if (res.status === 404) return null;
  if (!res.ok) {
    throw await readError(res, "Could not load this order");
  }
  const body: OrderResponse = await res.json();
  return body.data;
}

/** getLibrary lists the games the user owns, most recently acquired first. */
export async function getLibrary(): Promise<LibraryItem[]> {
  const res = await authedFetch("/library");
  if (!res.ok) {
    throw await readError(res, "Could not load your library");
  }
  const body: LibraryResponse = await res.json();
  return body.data;
}

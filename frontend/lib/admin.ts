import { authedFetch, jsonRequest, readError } from "@/lib/auth";
import type { Game } from "@/types/game";

/**
 * Admin game management. The API checks the role on every request, so these
 * calls are the real gate; hiding the admin UI from non-admins is only a
 * convenience. A 403 here means the signed-in user is not an admin, a 409 means
 * the slug is taken, and a 400 carries a validation message worth showing next
 * to the form.
 */

/**
 * AdminGameInput is the shape the admin endpoints accept. Genres and gallery
 * images are plain names and URLs; the API resolves genre names to rows and
 * creates any that do not exist yet.
 */
export type AdminGameInput = {
  title: string;
  slug: string;
  description: string;
  developer: string;
  publisher: string;
  /** YYYY-MM-DD, or "" to leave it unset. */
  release_date: string;
  price: number;
  /** Null removes the discount. */
  discount_price: number | null;
  /** RFC 3339 timestamp, or null for a discount that never expires on its own. */
  discount_ends_at: string | null;
  cover_image_url: string;
  genres: string[];
  images: string[];
};

type GameBody = { data: Game };

export async function createGame(input: AdminGameInput): Promise<Game> {
  const res = await authedFetch("/admin/games", jsonRequest("POST", input));
  if (!res.ok) {
    throw await readError(res, "Could not create the game");
  }
  const body: GameBody = await res.json();
  return body.data;
}

export async function updateGame(
  id: number,
  input: AdminGameInput,
): Promise<Game> {
  const res = await authedFetch(`/admin/games/${id}`, jsonRequest("PUT", input));
  if (!res.ok) {
    throw await readError(res, "Could not save the game");
  }
  const body: GameBody = await res.json();
  return body.data;
}

/**
 * deleteGame removes a game from the store. The API soft-deletes it, so past
 * orders and library entries that point at it keep working.
 */
export async function deleteGame(id: number): Promise<void> {
  const res = await authedFetch(`/admin/games/${id}`, { method: "DELETE" });
  if (!res.ok) {
    throw await readError(res, "Could not delete the game");
  }
}

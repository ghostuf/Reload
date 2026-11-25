import type { Game } from "@/types/game";

/**
 * Pricing display rules. The API is the source of truth for what a game costs:
 * the store list and detail endpoints send `effective_price`, and checkout
 * charges the same figure. Nothing here recomputes a price the API already
 * gave us.
 *
 * Games embedded in other responses (cart, wishlist) arrive without
 * `effective_price`, so for those alone currentPrice falls back to the API's
 * own rule: the discount price applies while set and not yet expired.
 */

type Priced = Pick<
  Game,
  "price" | "discount_price" | "discount_ends_at" | "effective_price"
>;

function discountStillRunning(game: Priced, now: number): boolean {
  if (game.discount_price == null) return false;
  if (game.discount_ends_at == null) return true;
  return new Date(game.discount_ends_at).getTime() > now;
}

/** currentPrice is what the game costs right now. */
export function currentPrice(game: Priced, now: number = Date.now()): number {
  if (game.effective_price != null) return game.effective_price;
  return discountStillRunning(game, now) ? game.discount_price! : game.price;
}

export type SaleInfo = {
  /** The list price, shown struck through. */
  original: number;
  /** What the game costs on sale. */
  sale: number;
  /** Whole-number percentage off, for the "-30%" badge. */
  percentOff: number;
};

/** saleInfo describes an active sale, or returns null when the game is at full price. */
export function saleInfo(game: Priced, now: number = Date.now()): SaleInfo | null {
  const sale = currentPrice(game, now);
  if (!(sale < game.price) || game.price <= 0) return null;

  return {
    original: game.price,
    sale,
    // At least 1%, so a tiny discount never reads as "-0%".
    percentOff: Math.max(1, Math.round((1 - sale / game.price) * 100)),
  };
}

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

/**
 * saleEndsLabel is the short countdown under a sale price, e.g. "Sale ends in
 * 2 days". It returns null when the sale has no end date or has already ended.
 */
export function saleEndsLabel(
  endsAt: string | null,
  now: number = Date.now(),
): string | null {
  if (!endsAt) return null;

  const left = new Date(endsAt).getTime() - now;
  if (!(left > 0)) return null;

  if (left < HOUR) return "Sale ends in less than an hour";
  if (left < DAY) {
    const hours = Math.floor(left / HOUR);
    return `Sale ends in ${hours} ${hours === 1 ? "hour" : "hours"}`;
  }
  const days = Math.floor(left / DAY);
  return `Sale ends in ${days} ${days === 1 ? "day" : "days"}`;
}

export function formatPrice(amount: number): string {
  return `Rs ${amount.toLocaleString("en-US", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  })}`;
}

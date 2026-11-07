import type { AdminGameInput } from "@/lib/admin";
import type { Game } from "@/types/game";

/**
 * GameFormValues is the admin form as typed: every field a string (or list of
 * strings) so half-typed numbers and dates survive until submit.
 */
export type GameFormValues = {
  title: string;
  slug: string;
  description: string;
  developer: string;
  publisher: string;
  /** YYYY-MM-DD from a date input, or "". */
  releaseDate: string;
  price: string;
  discountPrice: string;
  /** "YYYY-MM-DDTHH:mm" from a datetime-local input (local time), or "". */
  discountEndsAt: string;
  coverImageUrl: string;
  genres: string[];
  images: string[];
};

export type GameFormErrors = Partial<Record<keyof GameFormValues, string>>;

export const EMPTY_GAME_FORM: GameFormValues = {
  title: "",
  slug: "",
  description: "",
  developer: "",
  publisher: "",
  releaseDate: "",
  price: "",
  discountPrice: "",
  discountEndsAt: "",
  coverImageUrl: "",
  genres: [],
  images: [],
};

/** slugify turns a title into a URL slug: "Neon Drift 2!" -> "neon-drift-2". */
export function slugify(title: string): string {
  return title
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[̀-ͯ]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function pad(n: number) {
  return String(n).padStart(2, "0");
}

/** toLocalInput formats an ISO timestamp for a datetime-local input, in the viewer's time zone. */
function toLocalInput(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** valuesFromGame pre-fills the form for editing an existing game. */
export function valuesFromGame(game: Game): GameFormValues {
  // The API sends Go's zero time for a game with no release date.
  const release = game.release_date?.startsWith("0001-") ? "" : (game.release_date ?? "").slice(0, 10);

  return {
    title: game.title,
    slug: game.slug,
    description: game.description ?? "",
    developer: game.developer ?? "",
    publisher: game.publisher ?? "",
    releaseDate: release,
    price: String(game.price),
    discountPrice: game.discount_price == null ? "" : String(game.discount_price),
    discountEndsAt: game.discount_ends_at ? toLocalInput(game.discount_ends_at) : "",
    coverImageUrl: game.cover_image_url ?? "",
    genres: game.genres?.map((g) => g.name) ?? [],
    images: game.images?.map((i) => i.image_url) ?? [],
  };
}

const SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

function isHttpUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

/** A price field: digits with at most two decimals. */
function parsePrice(value: string): number | null {
  const trimmed = value.trim();
  if (!/^-?\d+(\.\d{1,2})?$/.test(trimmed)) return null;
  return Number(trimmed);
}

/**
 * validateGameForm checks what the browser can check before sending, and
 * builds the request body when everything passes. The API re-validates all of
 * it; this only saves a round trip for the obvious mistakes.
 */
export function validateGameForm(
  values: GameFormValues,
): { ok: true; input: AdminGameInput } | { ok: false; errors: GameFormErrors } {
  const errors: GameFormErrors = {};

  if (!values.title.trim()) errors.title = "Title is required.";

  const slug = values.slug.trim();
  if (!slug) errors.slug = "Slug is required.";
  else if (!SLUG_PATTERN.test(slug)) {
    errors.slug = "Use lowercase letters, numbers, and single hyphens.";
  }

  const cover = values.coverImageUrl.trim();
  if (!cover) errors.coverImageUrl = "Cover image URL is required.";
  else if (!isHttpUrl(cover)) errors.coverImageUrl = "Enter a full http(s) URL.";

  const price = parsePrice(values.price);
  if (values.price.trim() === "") errors.price = "Price is required.";
  else if (price === null) errors.price = "Enter a price like 19.99.";
  else if (price < 0) errors.price = "Price cannot be negative.";

  let discountPrice: number | null = null;
  if (values.discountPrice.trim() !== "") {
    discountPrice = parsePrice(values.discountPrice);
    if (discountPrice === null) errors.discountPrice = "Enter a price like 9.99.";
    else if (discountPrice < 0) errors.discountPrice = "Discount price cannot be negative.";
    else if (price !== null && price >= 0 && discountPrice >= price) {
      errors.discountPrice = "Discount price must be lower than the price.";
    }
  }

  let discountEndsAt: string | null = null;
  if (values.discountEndsAt) {
    const ends = new Date(values.discountEndsAt);
    if (Number.isNaN(ends.getTime())) errors.discountEndsAt = "Enter a valid date and time.";
    else if (values.discountPrice.trim() === "") {
      errors.discountEndsAt = "Set a discount price, or clear the end date.";
    } else discountEndsAt = ends.toISOString();
  }

  if (values.releaseDate && !/^\d{4}-\d{2}-\d{2}$/.test(values.releaseDate)) {
    errors.releaseDate = "Use the YYYY-MM-DD format.";
  }

  const images = values.images.map((u) => u.trim()).filter(Boolean);
  if (images.some((u) => !isHttpUrl(u))) {
    errors.images = "Every gallery image needs a full http(s) URL.";
  }

  if (Object.keys(errors).length > 0) return { ok: false, errors };

  return {
    ok: true,
    input: {
      title: values.title.trim(),
      slug,
      description: values.description.trim(),
      developer: values.developer.trim(),
      publisher: values.publisher.trim(),
      release_date: values.releaseDate,
      price: price!,
      discount_price: discountPrice,
      discount_ends_at: discountEndsAt,
      cover_image_url: cover,
      genres: values.genres,
      images,
    },
  };
}

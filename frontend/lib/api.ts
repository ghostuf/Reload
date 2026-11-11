import type { GameResponse, GamesResponse } from "@/types/game";

/**
 * The browser and the Next.js server never reach the API at the same
 * address once this runs in Docker: a browser needs the publicly reachable
 * URL (NEXT_PUBLIC_API_URL, inlined into the client bundle at build time),
 * while code running inside the frontend container -- a Server Component
 * doing its own fetch -- has to use the compose network's service name
 * instead (e.g. http://backend:8080), since "localhost" inside that
 * container is the frontend itself, not the backend.
 *
 * INTERNAL_API_URL deliberately has no NEXT_PUBLIC_ prefix so Next never
 * inlines it into the client bundle; it is read fresh from the real
 * environment at request time, which is exactly what a server-only variable
 * needs in standalone output.
 */
function resolveApiBaseUrl(): string {
  if (typeof window === "undefined") {
    return (
      process.env.INTERNAL_API_URL ??
      process.env.NEXT_PUBLIC_API_URL ??
      "http://localhost:8080"
    );
  }
  return process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
}

export const API_BASE_URL = resolveApiBaseUrl();

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status = 0) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export type ListGamesParams = {
  search?: string;
  genre?: string;
  /** Only games whose discount is currently active. */
  onSale?: boolean;
  sort?: string;
  page?: number;
  pageSize?: number;
};

export async function getGames(params: ListGamesParams): Promise<GamesResponse> {
  const query = new URLSearchParams();
  if (params.search) query.set("search", params.search);
  if (params.genre) query.set("genre", params.genre);
  if (params.onSale) query.set("on_sale", "true");
  if (params.sort) query.set("sort", params.sort);
  if (params.page) query.set("page", String(params.page));
  if (params.pageSize) query.set("page_size", String(params.pageSize));

  const res = await fetch(`${API_BASE_URL}/games?${query.toString()}`);
  if (!res.ok) {
    throw new ApiError(`Failed to fetch games (${res.status})`, res.status);
  }
  return res.json();
}

export async function getGame(id: string) {
  // A non-numeric id comes back 400; to a visitor it is just as missing.
  if (!/^\d+$/.test(id)) return null;
  const res = await fetch(`${API_BASE_URL}/games/${id}`);
  if (res.status === 404) return null;
  if (!res.ok) {
    throw new ApiError(`Failed to fetch game (${res.status})`, res.status);
  }
  const body: GameResponse = await res.json();
  return body.data;
}

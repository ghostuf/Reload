import type { GameResponse, GamesResponse } from "@/types/game";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {}

export type ListGamesParams = {
  search?: string;
  genre?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
};

export async function getGames(params: ListGamesParams): Promise<GamesResponse> {
  const query = new URLSearchParams();
  if (params.search) query.set("search", params.search);
  if (params.genre) query.set("genre", params.genre);
  if (params.sort) query.set("sort", params.sort);
  if (params.page) query.set("page", String(params.page));
  if (params.pageSize) query.set("page_size", String(params.pageSize));

  const res = await fetch(`${API_BASE_URL}/games?${query.toString()}`);
  if (!res.ok) {
    throw new ApiError(`Failed to fetch games (${res.status})`);
  }
  return res.json();
}

export async function getGame(id: string) {
  const res = await fetch(`${API_BASE_URL}/games/${id}`);
  if (res.status === 404) return null;
  if (!res.ok) {
    throw new ApiError(`Failed to fetch game (${res.status})`);
  }
  const body: GameResponse = await res.json();
  return body.data;
}

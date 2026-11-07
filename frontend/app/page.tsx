import { Suspense } from "react";
import GameFilters from "@/components/GameFilters";
import GameGrid, { GameGridSkeleton } from "@/components/GameGrid";
import GenreChips from "@/components/GenreChips";
import Pagination from "@/components/Pagination";
import RefreshButton from "@/components/RefreshButton";
import { getGames, type ListGamesParams } from "@/lib/api";
import type { GamesResponse } from "@/types/game";

type SearchParams = {
  search?: string;
  genre?: string;
  on_sale?: string;
  sort?: string;
  page?: string;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default async function Home({
  searchParams,
}: {
  searchParams: Promise<SearchParams>;
}) {
  const params = await searchParams;

  return (
    <main className="flex-1">
      <section className="mx-auto max-w-[1440px] px-4 pb-8 pt-12 sm:px-6 md:px-12 md:pt-24">
        <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
          Digital Game Store
        </p>
        <h1 className="mt-4 max-w-2xl text-4xl font-black leading-tight tracking-tight text-on-surface md:text-6xl">
          Reload your <span className="text-neon-lime">next adventure.</span>
        </h1>
        <p className="mt-4 max-w-xl text-on-surface-variant">
          Browse the catalog, search by title, and filter by genre to find
          what to play next.
        </p>
      </section>

      <section className="mx-auto max-w-[1440px] px-4 sm:px-6 md:px-12">
        <GameFilters />
      </section>

      <section className="mx-auto max-w-[1440px] px-4 pt-6 sm:px-6 md:px-12">
        <GenreChips active={params.genre} searchParams={params} />
      </section>

      <section className="mx-auto max-w-[1440px] px-4 py-10 sm:px-6 md:px-12">
        {/* Keyed on the query, so every new search or filter shows the skeleton
            instead of the previous results. */}
        <Suspense key={JSON.stringify(params)} fallback={<GameGridSkeleton />}>
          <GameResults params={params} />
        </Suspense>
      </section>
    </main>
  );
}

async function GameResults({ params }: { params: SearchParams }) {
  const page = Number(params.page) > 0 ? Number(params.page) : 1;

  const listParams: ListGamesParams = {
    search: params.search,
    genre: params.genre,
    onSale: params.on_sale === "true",
    sort: params.sort,
    page,
    pageSize: 12,
  };

  let result: GamesResponse | null = null;
  try {
    result = await getGames(listParams);
  } catch {
    result = null;
  }

  if (result === null) {
    return (
      <div
        role="alert"
        className="flex flex-col items-center rounded-game-card border border-white/5 bg-surface-container-low px-6 py-24 text-center"
      >
        <p className="text-lg font-semibold text-on-surface">
          Can&apos;t reach the Reload API
        </p>
        <p className="mt-2 text-sm text-on-surface-variant">
          Make sure the backend is running at {API_BASE_URL}.
        </p>
        <RefreshButton />
      </div>
    );
  }

  return (
    <>
      <GameGrid games={result.data} />
      <div className="pt-10">
        <Pagination
          page={result.pagination.page}
          totalPages={result.pagination.total_pages}
          searchParams={params}
        />
      </div>
    </>
  );
}

"use client";

import Image from "next/image";
import Link from "next/link";
import RequireAuth from "@/components/RequireAuth";
import { useLibrary } from "@/hooks/useLibrary";
import type { LibraryItem } from "@/types/order";
import ErrorState from "@/components/ErrorState";

export default function LibraryView() {
  return (
    <RequireAuth>
      <LibraryContents />
    </RequireAuth>
  );
}

function LibraryContents() {
  const library = useLibrary();
  const count = library.items.length;

  return (
    <main className="flex-1">
      <section className="mx-auto max-w-[1440px] px-4 sm:px-6 pb-8 pt-16 md:px-12 md:pt-24">
        <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
          Your Games
        </p>
        <h1 className="mt-4 text-4xl font-black tracking-tight text-on-surface md:text-5xl">
          Library
        </h1>
        {!library.loading && count > 0 && (
          <p className="mt-2 text-sm text-on-surface-variant">
            {count} {count === 1 ? "game" : "games"}
          </p>
        )}
      </section>

      <section className="mx-auto max-w-[1440px] px-4 sm:px-6 pb-20 md:px-12">
        {library.loading ? (
          <LibrarySkeleton />
        ) : library.error ? (
          <ErrorState
            title="Couldn't load your library"
            message={library.error}
            onRetry={library.retry}
          />
        ) : count === 0 ? (
          <EmptyLibrary />
        ) : (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-6 lg:grid-cols-4">
            {library.items.map((item, index) => (
              <LibraryCard
                key={item.id}
                item={item}
                preload={index === 0}
              />
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

/**
 * LibraryCard is the store card without its commerce: an owned game has no
 * price to show and nothing to add to a cart, so the tile is a plain link into
 * the game, with the date it was acquired in place of the price.
 */
function LibraryCard({
  item,
  preload,
}: {
  item: LibraryItem;
  preload: boolean;
}) {
  const { game } = item;

  return (
    <Link
      href={`/games/${item.game_id}`}
      className="group flex flex-col overflow-hidden rounded-game-card bg-surface-container-low transition-transform duration-300 hover:-translate-y-1"
    >
      <div className="relative aspect-[3/4] overflow-hidden">
        <Image
          src={game.cover_image_url}
          alt={game.title}
          fill
          sizes="(max-width: 768px) 50vw, (max-width: 1200px) 33vw, 25vw"
          className="object-cover transition-transform duration-500 group-hover:scale-105"
          preload={preload}
        />
        <span className="absolute left-3 top-3 rounded-full bg-neon-lime px-3 py-1 text-[11px] font-bold uppercase text-black">
          Owned
        </span>
      </div>
      <div className="flex flex-1 flex-col gap-2 p-4">
        <div className="flex flex-wrap gap-1.5">
          {game.genres?.slice(0, 2).map((genre) => (
            <span
              key={genre.id}
              className="rounded-full bg-white/5 px-2 py-0.5 text-[11px] uppercase tracking-wide text-on-surface-variant"
            >
              {genre.name}
            </span>
          ))}
        </div>
        <h3 className="line-clamp-1 font-bold text-on-surface transition-colors group-hover:text-electric-blue">
          {game.title}
        </h3>
        <p className="mt-auto pt-2 text-sm text-on-surface-variant">
          Added{" "}
          {new Date(item.acquired_at).toLocaleDateString(undefined, {
            year: "numeric",
            month: "short",
            day: "numeric",
          })}
        </p>
      </div>
    </Link>
  );
}

function LibraryIcon({ className = "h-6 w-6" }: { className?: string }) {
  return (
    <svg
      aria-hidden
      viewBox="0 0 24 24"
      className={className}
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M4 4.5h4v15H4zM10 4.5h4v15h-4z" />
      <path d="m16.3 5.6 3.6 14.2" />
    </svg>
  );
}

function EmptyLibrary() {
  return (
    <div className="flex flex-col items-center gap-3 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-24 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-full bg-white/5 text-on-surface-variant">
        <LibraryIcon />
      </span>
      <p className="text-lg font-semibold text-on-surface">
        Your library is empty
      </p>
      <p className="text-sm text-on-surface-variant">
        Games you buy land here, yours to keep.
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

function LibrarySkeleton({ tiles = 8 }: { tiles?: number }) {
  return (
    <div
      aria-hidden
      className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-6 lg:grid-cols-4"
    >
      {Array.from({ length: tiles }, (_, i) => (
        <div
          key={i}
          className="animate-pulse overflow-hidden rounded-game-card bg-surface-container-low"
        >
          <div className="aspect-[3/4] bg-white/5" />
          <div className="flex flex-col gap-2 p-4">
            <div className="h-3 w-1/3 rounded bg-white/5" />
            <div className="h-4 w-3/4 rounded bg-white/5" />
          </div>
        </div>
      ))}
    </div>
  );
}

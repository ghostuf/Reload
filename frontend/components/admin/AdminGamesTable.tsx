"use client";

import Link from "next/link";
import { useCallback, useState } from "react";
import { useLoad } from "@/hooks/useLoad";
import { getGames } from "@/lib/api";
import * as adminApi from "@/lib/admin";
import { formatPrice, saleInfo } from "@/lib/pricing";
import type { Game } from "@/types/game";

const PAGE_SIZE = 50;

/**
 * discountStatus summarises a game's discount for the table. "Expired" covers
 * a discount price that is still set but whose end date has passed — the store
 * already charges full price, but the admin may want to clear or extend it.
 */
export function discountStatus(game: Game, now: number = Date.now()) {
  if (game.discount_price == null) return { label: "—", tone: "none" as const };

  const sale = saleInfo(game, now);
  if (sale) {
    const ends = game.discount_ends_at
      ? ` · until ${new Date(game.discount_ends_at).toLocaleDateString(undefined, {
          month: "short",
          day: "numeric",
        })}`
      : "";
    return { label: `-${sale.percentOff}% · ${formatPrice(sale.sale)}${ends}`, tone: "active" as const };
  }

  const expired =
    game.discount_ends_at != null && new Date(game.discount_ends_at).getTime() <= now;
  return {
    label: expired ? "Expired" : "Not lower than price",
    tone: "inactive" as const,
  };
}

export default function AdminGamesTable() {
  const [page, setPage] = useState(1);
  const [version, setVersion] = useState(0);
  // Games removed in this session, hidden until the next reload confirms it.
  const [removed, setRemoved] = useState<ReadonlySet<number>>(() => new Set());
  const [deleting, setDeleting] = useState<number | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const load = useCallback(
    () => getGames({ page, pageSize: PAGE_SIZE }),
    // version forces a fresh read after "Try again".
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [page, version],
  );
  const { data, loading, error } = useLoad(load);

  async function handleDelete(game: Game) {
    if (!window.confirm(`Delete "${game.title}"? It will be removed from the store. Existing orders and libraries keep it.`)) {
      return;
    }
    setDeleting(game.id);
    setDeleteError(null);
    try {
      await adminApi.deleteGame(game.id);
      setRemoved((prev) => new Set(prev).add(game.id));
    } catch (err) {
      setDeleteError(
        `${game.title}: ${err instanceof Error && err.message ? err.message : "Could not delete the game"}`,
      );
    } finally {
      setDeleting(null);
    }
  }

  const games = data?.data.filter((g) => !removed.has(g.id)) ?? [];
  const totalPages = data?.pagination.total_pages ?? 1;

  return (
    <div className="mt-2">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-black tracking-tight text-on-surface">Games</h1>
          {data && (
            <p className="mt-1 text-sm text-on-surface-variant">
              {data.pagination.total - removed.size} in the store
            </p>
          )}
        </div>
        <Link
          href="/admin/games/new"
          className="rounded-xl bg-electric-blue px-4 py-2 text-sm font-bold text-white transition-transform active:scale-95"
        >
          New Game
        </Link>
      </div>

      {deleteError && (
        <p role="alert" className="mt-4 rounded-xl border border-red-400/30 bg-red-400/10 px-4 py-3 text-sm text-red-400">
          {deleteError}
        </p>
      )}

      <div className="mt-6">
        {error ? (
          <div className="rounded-game-card border border-white/5 bg-surface-container-low py-16 text-center">
            <p className="font-semibold text-on-surface">Couldn&apos;t load games</p>
            <p className="mt-1 text-sm text-on-surface-variant">{error}</p>
            <button
              type="button"
              onClick={() => setVersion((v) => v + 1)}
              className="mt-4 rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
            >
              Try again
            </button>
          </div>
        ) : (
          // The table keeps its columns on small screens and scrolls sideways.
          <div className="overflow-x-auto rounded-game-card border border-white/5 bg-surface-container-low">
            <table className="w-full min-w-[720px] text-left text-sm">
              <thead className="border-b border-white/10 text-xs uppercase tracking-wide text-on-surface-variant">
                <tr>
                  <th scope="col" className="px-4 py-3 font-semibold">Cover</th>
                  <th scope="col" className="px-4 py-3 font-semibold">Title</th>
                  <th scope="col" className="px-4 py-3 font-semibold">Price</th>
                  <th scope="col" className="px-4 py-3 font-semibold">Discount</th>
                  <th scope="col" className="px-4 py-3 text-right font-semibold">
                    <span className="sr-only">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {loading
                  ? Array.from({ length: 6 }, (_, i) => <SkeletonRow key={i} />)
                  : games.map((game) => (
                      <Row
                        key={game.id}
                        game={game}
                        deleting={deleting === game.id}
                        onDelete={() => handleDelete(game)}
                      />
                    ))}
                {!loading && games.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-4 py-12 text-center text-on-surface-variant">
                      No games yet.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {totalPages > 1 && (
        <div className="mt-6 flex items-center justify-center gap-4">
          <button
            type="button"
            onClick={() => setPage((p) => p - 1)}
            disabled={page <= 1 || loading}
            className="rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10 disabled:pointer-events-none disabled:opacity-30"
          >
            Previous
          </button>
          <span className="text-sm text-on-surface-variant">
            Page {page} of {totalPages}
          </span>
          <button
            type="button"
            onClick={() => setPage((p) => p + 1)}
            disabled={page >= totalPages || loading}
            className="rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10 disabled:pointer-events-none disabled:opacity-30"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

function Row({
  game,
  deleting,
  onDelete,
}: {
  game: Game;
  deleting: boolean;
  onDelete: () => void;
}) {
  const status = discountStatus(game);

  return (
    <tr className={deleting ? "opacity-50" : ""}>
      <td className="px-4 py-3">
        {/* A plain img: admins may enter covers from hosts next/image is not configured for. */}
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={game.cover_image_url}
          alt=""
          className="h-14 w-10 rounded-md bg-white/5 object-cover"
          loading="lazy"
        />
      </td>
      <td className="px-4 py-3">
        <Link
          href={`/games/${game.id}`}
          className="font-bold text-on-surface transition-colors hover:text-electric-blue"
        >
          {game.title}
        </Link>
        <p className="text-xs text-on-surface-variant">{game.slug}</p>
      </td>
      <td className="px-4 py-3 font-semibold text-on-surface">{formatPrice(game.price)}</td>
      <td className="px-4 py-3">
        <span
          className={
            status.tone === "active"
              ? "font-semibold text-neon-lime"
              : "text-on-surface-variant"
          }
        >
          {status.label}
        </span>
      </td>
      <td className="px-4 py-3">
        <div className="flex justify-end gap-2">
          <Link
            href={`/admin/games/${game.id}/edit`}
            className="rounded-xl border border-white/10 px-3 py-1.5 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
          >
            Edit
          </Link>
          <button
            type="button"
            onClick={onDelete}
            disabled={deleting}
            className="rounded-xl border border-red-400/30 px-3 py-1.5 text-sm font-semibold text-red-400 transition-colors hover:bg-red-400/10 disabled:opacity-50"
          >
            {deleting ? "Deleting…" : "Delete"}
          </button>
        </div>
      </td>
    </tr>
  );
}

function SkeletonRow() {
  return (
    <tr aria-hidden className="animate-pulse">
      <td className="px-4 py-3"><div className="h-14 w-10 rounded-md bg-white/5" /></td>
      <td className="px-4 py-3"><div className="h-4 w-40 rounded bg-white/5" /></td>
      <td className="px-4 py-3"><div className="h-4 w-14 rounded bg-white/5" /></td>
      <td className="px-4 py-3"><div className="h-4 w-24 rounded bg-white/5" /></td>
      <td className="px-4 py-3"><div className="ml-auto h-8 w-32 rounded-xl bg-white/5" /></td>
    </tr>
  );
}

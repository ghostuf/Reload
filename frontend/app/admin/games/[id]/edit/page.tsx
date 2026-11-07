"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useState } from "react";
import GameForm from "@/components/admin/GameForm";
import { useLoad } from "@/hooks/useLoad";
import { updateGame } from "@/lib/admin";
import { getGame } from "@/lib/api";
import { valuesFromGame } from "@/lib/gameForm";

export default function EditGamePage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const [attempt, setAttempt] = useState(0);

  const load = useCallback(
    () => getGame(id),
    // attempt forces a fresh read after "Try again".
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [id, attempt],
  );
  const { data: game, loading, error } = useLoad(load);

  return (
    <div className="mt-2">
      <Link
        href="/admin/games"
        className="text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface"
      >
        ← All games
      </Link>

      {loading ? (
        <div aria-hidden className="mt-6 flex max-w-3xl animate-pulse flex-col gap-4">
          <div className="h-9 w-64 rounded bg-white/5" />
          <div className="h-72 rounded-game-card bg-surface-container-low" />
          <div className="h-40 rounded-game-card bg-surface-container-low" />
        </div>
      ) : error ? (
        <div className="mt-6 max-w-3xl rounded-game-card border border-white/5 bg-surface-container-low py-16 text-center">
          <p className="font-semibold text-on-surface">Couldn&apos;t load this game</p>
          <p className="mt-1 text-sm text-on-surface-variant">{error}</p>
          <button
            type="button"
            onClick={() => setAttempt((a) => a + 1)}
            className="mt-4 rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
          >
            Try again
          </button>
        </div>
      ) : !game ? (
        <div className="mt-6 max-w-3xl rounded-game-card border border-white/5 bg-surface-container-low py-16 text-center">
          <p className="font-semibold text-on-surface">Game not found</p>
          <p className="mt-1 text-sm text-on-surface-variant">It may have been deleted.</p>
        </div>
      ) : (
        <>
          <h1 className="mt-3 text-3xl font-black tracking-tight text-on-surface">
            Edit {game.title}
          </h1>
          <GameForm
            key={game.id}
            isNew={false}
            initial={valuesFromGame(game)}
            onSubmit={async (input) => {
              await updateGame(game.id, input);
              router.push("/admin/games");
            }}
          />
        </>
      )}
    </div>
  );
}

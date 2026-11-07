import type { Game } from "@/types/game";
import GameCard from "./GameCard";

const GRID = "grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-6 lg:grid-cols-4";

export default function GameGrid({ games }: { games: Game[] }) {
  if (games.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-24 text-center">
        <p className="text-lg font-semibold text-on-surface">
          No games found
        </p>
        <p className="text-sm text-on-surface-variant">
          Try a different search term, genre, or filter.
        </p>
      </div>
    );
  }

  return (
    <div className={GRID}>
      {games.map((game, index) => (
        <GameCard key={game.id} game={game} preload={index === 0} />
      ))}
    </div>
  );
}

/** GameGridSkeleton holds the grid's place while games load. */
export function GameGridSkeleton({ tiles = 8 }: { tiles?: number }) {
  return (
    <div aria-hidden className={GRID}>
      {Array.from({ length: tiles }, (_, i) => (
        <div
          key={i}
          className="animate-pulse overflow-hidden rounded-game-card bg-surface-container-low"
        >
          <div className="aspect-[3/4] bg-white/5" />
          <div className="flex flex-col gap-2 p-4">
            <div className="h-3 w-1/3 rounded bg-white/5" />
            <div className="h-4 w-3/4 rounded bg-white/5" />
            <div className="mt-2 h-4 w-16 rounded bg-white/5" />
            <div className="mt-2 h-9 w-full rounded-xl bg-white/5" />
          </div>
        </div>
      ))}
    </div>
  );
}

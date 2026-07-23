import type { Game } from "@/types/game";
import GameCard from "./GameCard";

export default function GameGrid({ games }: { games: Game[] }) {
  if (games.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 rounded-game-card border border-white/5 bg-surface-container-low py-24 text-center">
        <p className="text-lg font-semibold text-on-surface">
          No games found
        </p>
        <p className="text-sm text-on-surface-variant">
          Try a different search term or genre.
        </p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-6 sm:grid-cols-3 lg:grid-cols-4">
      {games.map((game, index) => (
        <GameCard key={game.id} game={game} preload={index === 0} />
      ))}
    </div>
  );
}

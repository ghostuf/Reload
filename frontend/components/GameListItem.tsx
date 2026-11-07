import Image from "next/image";
import Link from "next/link";
import Price from "@/components/Price";
import type { Game } from "@/types/game";

/**
 * GameListItem is one row in the cart or wishlist: cover, title, price, and
 * whatever actions the list offers for it.
 */
export default function GameListItem({
  game,
  actions,
}: {
  game: Game;
  actions: React.ReactNode;
}) {
  return (
    <li className="flex gap-4 rounded-game-card border border-white/5 bg-surface-container-low p-3 sm:items-center sm:gap-5 sm:p-4">
      <Link
        href={`/games/${game.id}`}
        className="relative aspect-[3/4] w-20 shrink-0 overflow-hidden rounded-xl sm:w-24"
      >
        <Image
          src={game.cover_image_url}
          alt={game.title}
          fill
          sizes="96px"
          className="object-cover"
        />
      </Link>

      <div className="flex min-w-0 flex-1 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between sm:gap-6">
        <div className="min-w-0">
          <Link
            href={`/games/${game.id}`}
            className="line-clamp-2 font-bold text-on-surface transition-colors hover:text-electric-blue"
          >
            {game.title}
          </Link>
          <p className="mt-1 truncate text-sm text-on-surface-variant">
            {game.developer}
          </p>
          <div className="mt-2 flex items-center gap-2">
            <Price game={game} showBadge />
          </div>
        </div>

        <div className="flex shrink-0 flex-wrap items-center gap-3">
          {actions}
        </div>
      </div>
    </li>
  );
}

/** GameListSkeleton holds the list's place while it loads. */
export function GameListSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <ul aria-hidden className="flex flex-col gap-3">
      {Array.from({ length: rows }, (_, i) => (
        <li
          key={i}
          className="flex animate-pulse gap-4 rounded-game-card border border-white/5 bg-surface-container-low p-3 sm:gap-5 sm:p-4"
        >
          <div className="aspect-[3/4] w-20 shrink-0 rounded-xl bg-white/5 sm:w-24" />
          <div className="flex flex-1 flex-col gap-2 py-1">
            <div className="h-4 w-1/2 rounded bg-white/5" />
            <div className="h-3 w-1/3 rounded bg-white/5" />
            <div className="h-4 w-16 rounded bg-white/5" />
          </div>
        </li>
      ))}
    </ul>
  );
}

import Image from "next/image";
import Link from "next/link";
import { CartButton, WishlistButton } from "@/components/GameActions";
import Price, { DiscountBadge } from "@/components/Price";
import { saleInfo } from "@/lib/pricing";
import type { Game } from "@/types/game";

export default function GameCard({
  game,
  preload = false,
}: {
  game: Game;
  preload?: boolean;
}) {
  const sale = saleInfo(game);

  return (
    // The actions sit beside the link rather than inside it: a button nested in
    // an <a> is invalid markup, and its clicks would also follow the link.
    <div className="relative flex flex-col overflow-hidden rounded-game-card bg-surface-container-low transition-transform duration-300 hover:-translate-y-1">
      <Link href={`/games/${game.id}`} className="group flex flex-1 flex-col">
        <div className="relative aspect-[3/4] overflow-hidden">
          <Image
            src={game.cover_image_url}
            alt={game.title}
            fill
            sizes="(max-width: 768px) 50vw, (max-width: 1200px) 33vw, 25vw"
            className="object-cover transition-transform duration-500 group-hover:scale-105"
            preload={preload}
          />
          {sale && (
            <DiscountBadge
              percentOff={sale.percentOff}
              className="absolute left-3 top-3"
            />
          )}
        </div>
        <div className="flex flex-1 flex-col gap-2 p-4 pb-3">
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
          <div className="mt-auto flex items-center gap-2 pt-2">
            <Price game={game} />
          </div>
        </div>
      </Link>
      <div className="absolute right-3 top-3">
        <WishlistButton gameId={game.id} />
      </div>
      <div className="px-4 pb-4">
        <CartButton gameId={game.id} />
      </div>
    </div>
  );
}

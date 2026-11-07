import { currentPrice, formatPrice, saleInfo } from "@/lib/pricing";
import type { Game } from "@/types/game";

type Size = "sm" | "lg";

const STYLES: Record<Size, { original: string; current: string; badge: string }> = {
  sm: {
    original: "text-sm",
    current: "font-bold",
    badge: "px-2 py-0.5 text-[11px]",
  },
  lg: {
    original: "text-lg",
    current: "text-2xl font-bold",
    badge: "px-2.5 py-1 text-xs",
  },
};

/**
 * Price shows what a game costs. During a sale it shows the list price struck
 * through beside the sale price; `showBadge` adds the "-30%" chip inline, for
 * places that do not already show it over the cover art.
 */
export default function Price({
  game,
  size = "sm",
  showBadge = false,
}: {
  game: Game;
  size?: Size;
  showBadge?: boolean;
}) {
  const sale = saleInfo(game);
  const styles = STYLES[size];

  if (!sale) {
    return (
      <span className={`${styles.current} text-on-surface`}>
        {formatPrice(currentPrice(game))}
      </span>
    );
  }

  return (
    <span className="flex flex-wrap items-center gap-2">
      {showBadge && <DiscountBadge percentOff={sale.percentOff} size={size} />}
      <span className={`${styles.original} text-on-surface-variant line-through`}>
        <span className="sr-only">Was </span>
        {formatPrice(sale.original)}
      </span>
      <span className={`${styles.current} text-on-surface`}>
        <span className="sr-only">Now </span>
        {formatPrice(sale.sale)}
      </span>
    </span>
  );
}

export function DiscountBadge({
  percentOff,
  size = "sm",
  className = "",
}: {
  percentOff: number;
  size?: Size;
  className?: string;
}) {
  return (
    <span
      className={`rounded-full bg-neon-lime font-bold uppercase text-black ${STYLES[size].badge} ${className}`}
    >
      -{percentOff}%
    </span>
  );
}

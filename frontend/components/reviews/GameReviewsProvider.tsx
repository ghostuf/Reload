"use client";

import { createContext, useCallback, useContext, useMemo, useState } from "react";
import { getGame } from "@/lib/api";
import { Stars } from "@/components/reviews/Stars";

export type ReviewSummary = {
  /** Null when there are no reviews yet. */
  average: number | null;
  count: number;
};

type GameReviewsContextValue = {
  summary: ReviewSummary;
  /** Re-reads the average and count from the API after a review changes. */
  refreshSummary: () => Promise<void>;
};

const GameReviewsContext = createContext<GameReviewsContextValue | null>(null);

/**
 * GameReviewsProvider shares one game's rating summary between the stars near
 * the title and the reviews section further down, so posting a review updates
 * both without reloading the page. The figures always come from the API.
 */
export default function GameReviewsProvider({
  gameId,
  initial,
  children,
}: {
  gameId: number;
  initial: ReviewSummary;
  children: React.ReactNode;
}) {
  const [summary, setSummary] = useState(initial);

  const refreshSummary = useCallback(async () => {
    try {
      const game = await getGame(String(gameId));
      if (game) {
        setSummary({ average: game.average_rating, count: game.review_count });
      }
    } catch {
      // The review itself was saved; a stale average is not worth an error.
    }
  }, [gameId]);

  const value = useMemo(
    () => ({ summary, refreshSummary }),
    [summary, refreshSummary],
  );

  return (
    <GameReviewsContext.Provider value={value}>{children}</GameReviewsContext.Provider>
  );
}

export function useGameReviews(): GameReviewsContextValue {
  const ctx = useContext(GameReviewsContext);
  if (!ctx) {
    throw new Error("useGameReviews must be used within a GameReviewsProvider");
  }
  return ctx;
}

/** RatingSummary is the average stars and review count shown by the title. */
export function RatingSummary() {
  const { summary } = useGameReviews();

  if (summary.count === 0 || summary.average == null) {
    return (
      <a
        href="#reviews"
        className="text-sm text-on-surface-variant transition-colors hover:text-on-surface"
      >
        No reviews yet
      </a>
    );
  }

  return (
    <a href="#reviews" className="group inline-flex items-center gap-2 text-sm">
      <Stars rating={summary.average} />
      <span className="font-bold text-on-surface">{summary.average.toFixed(1)}</span>
      <span className="text-on-surface-variant transition-colors group-hover:text-on-surface">
        ({summary.count} {summary.count === 1 ? "review" : "reviews"})
      </span>
    </a>
  );
}

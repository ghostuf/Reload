"use client";

import { useCallback, useState } from "react";
import ReviewComposer, { composerState } from "@/components/reviews/ReviewComposer";
import ReviewItem from "@/components/reviews/ReviewItem";
import { useGameReviews } from "@/components/reviews/GameReviewsProvider";
import { useAuth } from "@/hooks/useAuth";
import { useLibrary } from "@/hooks/useLibrary";
import { useLoad } from "@/hooks/useLoad";
import * as reviewsApi from "@/lib/reviews";
import type { Review, ReviewInput, SavedReview } from "@/types/review";

const PAGE_SIZE = 5;

/**
 * Own-review tracking. The list endpoint does not say who wrote each review,
 * so a review is recognised as the user's own when it was posted in this
 * session, or when the API includes `user_id` on list entries. `override` holds
 * what this session knows; undefined means "nothing yet, go by the list".
 */
type Override = { review: Review | null } | undefined;

export default function ReviewsSection({ gameId }: { gameId: number }) {
  const { user, loading: authLoading } = useAuth();
  const library = useLibrary();
  const { summary, refreshSummary } = useGameReviews();

  const [page, setPage] = useState(1);
  // Bumped after every write so the list re-reads from the API.
  const [version, setVersion] = useState(0);
  const [override, setOverride] = useState<Override>(undefined);

  const load = useCallback(
    () => reviewsApi.getReviews(gameId, { page, pageSize: PAGE_SIZE }),
    // version is a deliberate dependency: a new value forces a fresh read.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [gameId, page, version],
  );
  const { data, loading, error } = useLoad(load);

  const fromList =
    user && data ? (data.data.find((r) => r.user_id === user.id) ?? null) : null;
  const ownReview = override !== undefined ? override.review : fromList;

  const withName = useCallback(
    (saved: SavedReview): Review => ({
      ...saved,
      user_id: user?.id,
      display_name: user?.display_name ?? "",
    }),
    [user],
  );

  const afterWrite = useCallback(async () => {
    setVersion((v) => v + 1);
    await refreshSummary();
  }, [refreshSummary]);

  async function handleCreate(input: ReviewInput) {
    const saved = await reviewsApi.createReview(gameId, input);
    setOverride({ review: withName(saved) });
    setPage(1);
    await afterWrite();
  }

  async function handleUpdate(input: ReviewInput) {
    if (!ownReview) return;
    const saved = await reviewsApi.updateReview(ownReview.id, input);
    setOverride({ review: withName(saved) });
    await afterWrite();
  }

  async function handleDelete() {
    if (!ownReview) return;
    await reviewsApi.deleteReview(ownReview.id);
    setOverride({ review: null });
    await afterWrite();
  }

  const state = composerState({
    authLoading,
    signedIn: user !== null,
    libraryLoading: library.loading,
    owns: library.has(gameId),
    hasOwnReview: ownReview !== null,
  });

  const totalPages = data?.pagination.total_pages ?? 1;

  return (
    <section id="reviews" aria-labelledby="reviews-heading" className="scroll-mt-24">
      <h2 id="reviews-heading" className="text-xl font-bold text-on-surface">
        Reviews{" "}
        <span className="text-base font-semibold text-on-surface-variant">
          ({summary.count})
        </span>
      </h2>

      <div className="mt-4">
        <ReviewComposer
          // Remount on a different review so edit state never carries over.
          key={ownReview?.id ?? "none"}
          state={state}
          gameId={gameId}
          ownReview={ownReview}
          onCreate={handleCreate}
          onUpdate={handleUpdate}
          onDelete={handleDelete}
        />
      </div>

      <div className="mt-6">
        {error ? (
          <div className="rounded-game-card border border-white/5 bg-surface-container-low px-5 py-6 text-center">
            <p className="text-sm text-on-surface-variant">{error}</p>
            <button
              type="button"
              onClick={() => setVersion((v) => v + 1)}
              className="mt-3 rounded-xl border border-white/10 px-4 py-1.5 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
            >
              Try again
            </button>
          </div>
        ) : loading && !data ? (
          <ReviewListSkeleton />
        ) : data && data.data.length === 0 ? (
          <p className="text-sm text-on-surface-variant">
            No reviews yet.
          </p>
        ) : (
          <ul className={`flex flex-col gap-3 ${loading ? "opacity-60" : ""}`}>
            {data?.data.map((review) => (
              <ReviewItem key={review.id} review={review} />
            ))}
          </ul>
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
    </section>
  );
}

function ReviewListSkeleton() {
  return (
    <ul aria-hidden className="flex flex-col gap-3">
      {[0, 1, 2].map((i) => (
        <li key={i} className="animate-pulse rounded-game-card border border-white/5 bg-surface-container-low p-5">
          <div className="h-4 w-1/3 rounded bg-white/5" />
          <div className="mt-3 h-3 w-full rounded bg-white/5" />
          <div className="mt-2 h-3 w-2/3 rounded bg-white/5" />
        </li>
      ))}
    </ul>
  );
}

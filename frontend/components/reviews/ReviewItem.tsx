import { Stars } from "@/components/reviews/Stars";
import type { Review } from "@/types/review";

/** ReviewItem is one review: who wrote it, the stars, the date, and the text. */
export default function ReviewItem({
  review,
  bare = false,
}: {
  review: Review;
  /** Drop the card styling, for when the review sits inside another card. */
  bare?: boolean;
}) {
  const date = new Date(review.created_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
  const edited = review.updated_at !== review.created_at;

  const content = (
    <>
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <span className="font-bold text-on-surface">
          {review.display_name || "Anonymous"}
        </span>
        <Stars rating={review.rating} size={14} />
        <span className="text-xs text-on-surface-variant">
          {date}
          {edited && " · edited"}
        </span>
      </div>
      <p className="mt-2 whitespace-pre-line break-words leading-relaxed text-on-surface-variant">
        {review.body}
      </p>
    </>
  );

  if (bare) return <div>{content}</div>;

  return (
    <li className="rounded-game-card border border-white/5 bg-surface-container-low p-5">
      {content}
    </li>
  );
}

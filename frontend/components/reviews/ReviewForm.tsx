"use client";

import { useState } from "react";
import { StarInput } from "@/components/reviews/Stars";
import type { ReviewInput } from "@/types/review";

/** Matches the API's limit, so the counter warns before the server rejects. */
export const MAX_REVIEW_LENGTH = 4000;

/**
 * ReviewForm writes or edits a review. It checks the obvious things before
 * sending (a rating and some text), and shows whatever the API rejects with
 * inline, since only the API knows about ownership and duplicates.
 */
export default function ReviewForm({
  initial,
  submitLabel,
  onSubmit,
  onCancel,
}: {
  initial?: ReviewInput;
  submitLabel: string;
  /** Throws to report a failure; its message is shown under the form. */
  onSubmit: (input: ReviewInput) => Promise<void>;
  onCancel?: () => void;
}) {
  const [rating, setRating] = useState(initial?.rating ?? 0);
  const [body, setBody] = useState(initial?.body ?? "");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (rating < 1) {
      setError("Pick a rating from 1 to 5 stars.");
      return;
    }
    if (!body.trim()) {
      setError("Write a few words about the game.");
      return;
    }
    if (body.length > MAX_REVIEW_LENGTH) {
      setError(`Keep it under ${MAX_REVIEW_LENGTH} characters.`);
      return;
    }

    setError(null);
    setSubmitting(true);
    try {
      await onSubmit({ rating, body: body.trim() });
    } catch (err) {
      setError(err instanceof Error && err.message ? err.message : "Something went wrong");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      aria-label="Review form"
      className="flex flex-col gap-4 rounded-game-card border border-white/10 bg-surface-container-low p-5"
    >
      <StarInput value={rating} onChange={setRating} disabled={submitting} />
      <div>
        <label htmlFor="review-body" className="sr-only">
          Your review
        </label>
        <textarea
          id="review-body"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          disabled={submitting}
          rows={4}
          placeholder="What did you think?"
          className="w-full resize-y rounded-2xl border border-white/10 bg-surface-container-lowest px-4 py-3 text-on-surface outline-none placeholder:text-on-surface-variant focus:border-electric-blue"
        />
        <p
          className={`mt-1 text-right text-xs ${
            body.length > MAX_REVIEW_LENGTH ? "text-red-400" : "text-on-surface-variant"
          }`}
        >
          {body.length}/{MAX_REVIEW_LENGTH}
        </p>
      </div>

      {error && (
        <p role="alert" className="text-sm text-red-400">
          {error}
        </p>
      )}

      <div className="flex flex-wrap gap-3">
        <button
          type="submit"
          disabled={submitting}
          className="rounded-xl bg-electric-blue px-5 py-2 text-sm font-bold text-white transition-transform active:scale-95 disabled:opacity-50"
        >
          {submitting ? "Saving…" : submitLabel}
        </button>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={submitting}
            className="rounded-xl border border-white/10 px-5 py-2 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
          >
            Cancel
          </button>
        )}
      </div>
    </form>
  );
}

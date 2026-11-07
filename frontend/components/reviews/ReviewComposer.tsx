"use client";

import Link from "next/link";
import { useState } from "react";
import ReviewForm from "@/components/reviews/ReviewForm";
import ReviewItem from "@/components/reviews/ReviewItem";
import type { Review, ReviewInput } from "@/types/review";

export type ComposerState =
  | "loading"
  | "logged-out"
  | "not-owner"
  | "write"
  | "own-review";

/**
 * composerState decides what the review area shows. Kept separate from the
 * markup so the rules read in one place: signed-out visitors are asked to log
 * in, signed-in visitors who do not own the game are told to buy it, owners
 * see their review if they have one and the form if not.
 */
export function composerState({
  authLoading,
  signedIn,
  libraryLoading,
  owns,
  hasOwnReview,
}: {
  authLoading: boolean;
  signedIn: boolean;
  libraryLoading: boolean;
  owns: boolean;
  hasOwnReview: boolean;
}): ComposerState {
  if (authLoading) return "loading";
  if (!signedIn) return "logged-out";
  if (libraryLoading) return "loading";
  if (!owns) return "not-owner";
  return hasOwnReview ? "own-review" : "write";
}

export default function ReviewComposer({
  state,
  gameId,
  ownReview,
  onCreate,
  onUpdate,
  onDelete,
}: {
  state: ComposerState;
  gameId: number;
  ownReview: Review | null;
  onCreate: (input: ReviewInput) => Promise<void>;
  onUpdate: (input: ReviewInput) => Promise<void>;
  onDelete: () => Promise<void>;
}) {
  const [editing, setEditing] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  switch (state) {
    case "loading":
      return <div aria-hidden className="h-24 animate-pulse rounded-game-card bg-surface-container-low" />;

    case "logged-out":
      return (
        <Notice>
          <Link
            href={`/login?next=${encodeURIComponent(`/games/${gameId}`)}`}
            className="font-bold text-electric-blue hover:underline"
          >
            Log in to review
          </Link>{" "}
          this game.
        </Notice>
      );

    case "not-owner":
      return <Notice>Buy this game to leave a review.</Notice>;

    case "write":
      return <ReviewForm submitLabel="Post review" onSubmit={onCreate} />;

    case "own-review": {
      if (!ownReview) return null;
      if (editing) {
        return (
          <ReviewForm
            initial={{ rating: ownReview.rating, body: ownReview.body }}
            submitLabel="Save changes"
            onSubmit={async (input) => {
              await onUpdate(input);
              setEditing(false);
            }}
            onCancel={() => setEditing(false)}
          />
        );
      }

      async function handleDelete() {
        if (!window.confirm("Delete your review? This can't be undone.")) return;
        setDeleting(true);
        setDeleteError(null);
        try {
          await onDelete();
        } catch (err) {
          setDeleteError(err instanceof Error && err.message ? err.message : "Could not delete your review");
        } finally {
          setDeleting(false);
        }
      }

      return (
        <div className="rounded-game-card border border-electric-blue/40 bg-surface-container-low p-5">
          <p className="mb-3 text-xs font-bold uppercase tracking-widest text-electric-blue">
            Your review
          </p>
          <ReviewItem review={ownReview} bare />
          <div className="mt-4 flex flex-wrap gap-3">
            <button
              type="button"
              onClick={() => setEditing(true)}
              disabled={deleting}
              className="rounded-xl border border-white/10 px-4 py-1.5 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
            >
              Edit
            </button>
            <button
              type="button"
              onClick={handleDelete}
              disabled={deleting}
              className="rounded-xl border border-red-400/30 px-4 py-1.5 text-sm font-semibold text-red-400 transition-colors hover:bg-red-400/10 disabled:opacity-50"
            >
              {deleting ? "Deleting…" : "Delete"}
            </button>
          </div>
          {deleteError && (
            <p role="alert" className="mt-3 text-sm text-red-400">
              {deleteError}
            </p>
          )}
        </div>
      );
    }
  }
}

function Notice({ children }: { children: React.ReactNode }) {
  return (
    <p className="rounded-game-card border border-white/5 bg-surface-container-low px-5 py-4 text-sm text-on-surface-variant">
      {children}
    </p>
  );
}

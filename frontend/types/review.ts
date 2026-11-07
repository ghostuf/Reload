import type { Pagination } from "@/types/game";

/** Review is one entry in a game's public review list. */
export type Review = {
  id: number;
  game_id: number;
  /**
   * The author. The list endpoint does not send this yet; when it does, the
   * page can recognise the signed-in user's review on load.
   */
  user_id?: number;
  rating: number;
  body: string;
  display_name: string;
  created_at: string;
  updated_at: string;
};

/**
 * SavedReview is what creating or editing a review returns. The API leaves out
 * the reviewer's display name here, so callers fill it in from the signed-in
 * user before showing it.
 */
export type SavedReview = Omit<Review, "display_name">;

export type ReviewInput = {
  rating: number;
  body: string;
};

export type ReviewsResponse = {
  data: Review[];
  pagination: Pagination;
};

export type SavedReviewResponse = {
  data: SavedReview;
};

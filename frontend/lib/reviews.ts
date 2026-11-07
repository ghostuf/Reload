import { API_BASE_URL, ApiError } from "@/lib/api";
import { authedFetch, jsonRequest, readError } from "@/lib/auth";
import type {
  ReviewInput,
  ReviewsResponse,
  SavedReview,
  SavedReviewResponse,
} from "@/types/review";

/**
 * Reading reviews is public, so the list is a plain fetch. Writing one needs
 * the session, so those calls go through authedFetch. Failures carry the API's
 * own message and status: 403 means the user does not own the game (or the
 * review), 409 means they already reviewed it, and 400 is a validation error
 * worth showing next to the form.
 */

export type ListReviewsParams = {
  page?: number;
  pageSize?: number;
};

export async function getReviews(
  gameId: number | string,
  params: ListReviewsParams = {},
): Promise<ReviewsResponse> {
  const query = new URLSearchParams();
  if (params.page) query.set("page", String(params.page));
  if (params.pageSize) query.set("page_size", String(params.pageSize));

  const res = await fetch(
    `${API_BASE_URL}/games/${gameId}/reviews?${query.toString()}`,
  );
  if (!res.ok) {
    throw new ApiError(`Failed to fetch reviews (${res.status})`, res.status);
  }
  return res.json();
}

export async function createReview(
  gameId: number,
  input: ReviewInput,
): Promise<SavedReview> {
  const res = await authedFetch(
    `/games/${gameId}/reviews`,
    jsonRequest("POST", input),
  );
  if (!res.ok) {
    throw await readError(res, "Could not post your review");
  }
  const body: SavedReviewResponse = await res.json();
  return body.data;
}

export async function updateReview(
  reviewId: number,
  input: ReviewInput,
): Promise<SavedReview> {
  const res = await authedFetch(`/reviews/${reviewId}`, jsonRequest("PUT", input));
  if (!res.ok) {
    throw await readError(res, "Could not update your review");
  }
  const body: SavedReviewResponse = await res.json();
  return body.data;
}

export async function deleteReview(reviewId: number): Promise<void> {
  const res = await authedFetch(`/reviews/${reviewId}`, { method: "DELETE" });
  if (!res.ok) {
    throw await readError(res, "Could not delete your review");
  }
}

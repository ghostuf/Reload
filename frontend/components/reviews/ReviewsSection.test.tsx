import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ReviewsSection from "@/components/reviews/ReviewsSection";
import { useAuth } from "@/hooks/useAuth";
import { useLibrary } from "@/hooks/useLibrary";
import { getReviews } from "@/lib/reviews";
import { makeUser } from "@/test/fixtures";
import type { AuthContextValue } from "@/components/AuthProvider";
import type { LibraryContextValue } from "@/components/ShopProvider";
import type { User } from "@/types/user";

vi.mock("@/hooks/useAuth", () => ({ useAuth: vi.fn() }));
vi.mock("@/hooks/useLibrary", () => ({ useLibrary: vi.fn() }));
vi.mock("@/lib/reviews", () => ({
  getReviews: vi.fn(),
  createReview: vi.fn(),
  updateReview: vi.fn(),
  deleteReview: vi.fn(),
}));
vi.mock("@/components/reviews/GameReviewsProvider", () => ({
  useGameReviews: () => ({
    summary: { average: null, count: 0 },
    refreshSummary: vi.fn(),
  }),
}));

const GAME_ID = 42;

function signIn(user: User | null) {
  vi.mocked(useAuth).mockReturnValue({
    user,
    loading: false,
    isAdmin: user?.role === "admin",
    signIn: vi.fn(),
    signUp: vi.fn(),
    signOut: vi.fn(),
  } satisfies AuthContextValue);
}

function own(gameIds: number[]) {
  vi.mocked(useLibrary).mockReturnValue({
    items: [],
    loading: false,
    error: null,
    has: (id: number) => gameIds.includes(id),
    retry: vi.fn(),
    refresh: vi.fn(),
  } satisfies LibraryContextValue);
}

beforeEach(() => {
  vi.mocked(getReviews).mockResolvedValue({
    data: [],
    pagination: { page: 1, page_size: 5, total: 0, total_pages: 1 },
  });
});

async function renderSection() {
  render(<ReviewsSection gameId={GAME_ID} />);
  // Let the review list settle so nothing updates after the assertions.
  await waitFor(() => expect(screen.getByText("No reviews yet.")).toBeInTheDocument());
}

describe("review form visibility", () => {
  it("shows the form to a signed-in owner", async () => {
    signIn(makeUser());
    own([GAME_ID]);
    await renderSection();

    expect(screen.getByRole("form", { name: "Review form" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Post review" })).toBeInTheDocument();
    expect(screen.queryByText(/Buy this game/)).not.toBeInTheDocument();
  });

  it("tells a signed-in non-owner to buy the game instead", async () => {
    signIn(makeUser());
    own([]);
    await renderSection();

    expect(screen.getByText("Buy this game to leave a review.")).toBeInTheDocument();
    expect(screen.queryByRole("form", { name: "Review form" })).not.toBeInTheDocument();
  });

  it("asks a signed-out visitor to log in, and brings them back afterwards", async () => {
    signIn(null);
    own([]);
    await renderSection();

    const link = screen.getByRole("link", { name: "Log in to review" });
    expect(link).toHaveAttribute("href", `/login?next=${encodeURIComponent(`/games/${GAME_ID}`)}`);
    expect(screen.queryByRole("form", { name: "Review form" })).not.toBeInTheDocument();
  });
});

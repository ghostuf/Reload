import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import GameCard from "@/components/GameCard";
import { makeGame } from "@/test/fixtures";

// The cart and wishlist buttons need the shop and auth providers; they are not
// what these tests are about.
vi.mock("@/components/GameActions", () => ({
  CartButton: () => null,
  WishlistButton: () => null,
}));

vi.mock("next/image", () => ({
  // eslint-disable-next-line @next/next/no-img-element
  default: ({ src, alt }: { src: string; alt: string }) => <img src={src} alt={alt} />,
}));

describe("GameCard pricing", () => {
  it("shows the original price struck through, the sale price, and the discount badge", () => {
    render(
      <GameCard
        game={makeGame({ price: 30, discount_price: 21, effective_price: 21 })}
      />,
    );

    const original = screen.getByText("Rs 30");
    expect(original).toHaveClass("line-through");
    expect(screen.getByText("Rs 21")).toBeInTheDocument();
    expect(screen.getByText("-30%")).toBeInTheDocument();
  });

  it("uses the effective price from the API rather than recomputing it", () => {
    // discount_price says 10, but the API has decided the game costs 25 (for
    // example, a discount that just expired). The card must show the API's figure.
    render(
      <GameCard
        game={makeGame({ price: 50, discount_price: 10, effective_price: 25 })}
      />,
    );

    expect(screen.getByText("Rs 25")).toBeInTheDocument();
    expect(screen.queryByText("Rs 10")).not.toBeInTheDocument();
    expect(screen.getByText("-50%")).toBeInTheDocument();
  });

  it("shows a single price and no badge when the game is not on sale", () => {
    render(
      <GameCard
        game={makeGame({
          price: 30,
          discount_price: 20,
          // Expired: the API charges full price again.
          discount_ends_at: "2020-01-01T00:00:00Z",
          effective_price: 30,
        })}
      />,
    );

    const price = screen.getByText("Rs 30");
    expect(price).not.toHaveClass("line-through");
    expect(screen.queryByText(/^-\d+%$/)).not.toBeInTheDocument();
  });
});

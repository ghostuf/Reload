import type { Game } from "@/types/game";
import type { User } from "@/types/user";

export function makeGame(overrides: Partial<Game> = {}): Game {
  return {
    id: 1,
    title: "Neon Drift",
    slug: "neon-drift",
    description: "Arcade racing after dark.",
    developer: "Night Shift",
    publisher: "Reload Games",
    release_date: "2025-03-14T00:00:00Z",
    price: 29.99,
    discount_price: null,
    discount_ends_at: null,
    effective_price: 29.99,
    cover_image_url: "https://picsum.photos/seed/neon/600/800",
    genres: [{ id: 1, name: "Racing" }],
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
    ...overrides,
  };
}

export function makeUser(overrides: Partial<User> = {}): User {
  return {
    id: 7,
    email: "player@example.com",
    display_name: "Player One",
    role: "user",
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
    ...overrides,
  };
}

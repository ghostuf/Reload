import type { Game } from "@/types/game";

export type CartItem = {
  id: number;
  user_id: number;
  game_id: number;
  game: Game;
  created_at: string;
};

export type WishlistItem = {
  id: number;
  user_id: number;
  game_id: number;
  game: Game;
  created_at: string;
};

export type CartResponse = {
  data: CartItem[];
  total: number;
};

export type CartItemResponse = {
  data: CartItem;
};

export type WishlistResponse = {
  data: WishlistItem[];
};

export type WishlistItemResponse = {
  data: WishlistItem;
};

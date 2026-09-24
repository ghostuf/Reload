import type { Game } from "@/types/game";

/**
 * price_at_purchase is what the game cost when the order was placed, snapshot
 * by the API. It is deliberately not read off `game`, whose price may have
 * changed since.
 */
export type OrderItem = {
  id: number;
  order_id: number;
  game_id: number;
  game: Game;
  price_at_purchase: number;
};

export type Order = {
  id: number;
  user_id: number;
  total_price: number;
  status: string;
  items?: OrderItem[];
  created_at: string;
};

/**
 * OrderSummary is the lighter shape GET /orders returns per order: enough for
 * a history row, without every line item's game details.
 */
export type OrderSummary = {
  id: number;
  total_price: number;
  item_count: number;
  created_at: string;
};

export type LibraryItem = {
  id: number;
  user_id: number;
  game_id: number;
  game: Game;
  acquired_at: string;
};

export type OrderResponse = {
  data: Order;
};

export type OrdersResponse = {
  data: OrderSummary[];
};

export type LibraryResponse = {
  data: LibraryItem[];
};

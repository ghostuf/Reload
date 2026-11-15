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

/**
 * Payment is one payment attempt against an order -- present only for
 * eSewa-checkout orders; the mock checkout never creates one, so this is
 * absent on those.
 */
export type Payment = {
  id: number;
  order_id: number;
  provider: string;
  transaction_uuid: string;
  amount: number;
  status: string;
  provider_ref_id: string | null;
  created_at: string;
  updated_at: string;
};

export type Order = {
  id: number;
  user_id: number;
  total_price: number;
  status: string;
  items?: OrderItem[];
  payment?: Payment | null;
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
  status: string;
  payment_provider?: string;
  payment_status?: string;
  provider_ref_id?: string | null;
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

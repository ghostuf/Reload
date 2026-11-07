export type Genre = {
  id: number;
  name: string;
};

export type GameImage = {
  id: number;
  game_id: number;
  image_url: string;
  sort_order: number;
};

export type Game = {
  id: number;
  title: string;
  slug: string;
  description: string;
  developer: string;
  publisher: string;
  release_date: string;
  price: number;
  /** Set by an admin; only applies while the discount is active. */
  discount_price: number | null;
  /** When the discount stops applying. Null means it never expires on its own. */
  discount_ends_at: string | null;
  /**
   * What a shopper actually pays, computed by the API. Present on the list and
   * detail endpoints; embedded games (cart, orders, library) may omit it.
   */
  effective_price?: number;
  cover_image_url: string;
  genres?: Genre[];
  images?: GameImage[];
  created_at: string;
  updated_at: string;
};

/** GameDetail is what GET /games/:id returns: a game plus its rating summary. */
export type GameDetail = Game & {
  effective_price: number;
  /** Null when the game has no reviews yet. */
  average_rating: number | null;
  review_count: number;
};

export type Pagination = {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type GamesResponse = {
  data: Game[];
  pagination: Pagination;
};

export type GameResponse = {
  data: GameDetail;
};

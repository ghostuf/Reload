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
  discount_price: number | null;
  cover_image_url: string;
  genres?: Genre[];
  images?: GameImage[];
  created_at: string;
  updated_at: string;
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
  data: Game;
};

"use client";

import { useContext } from "react";
import {
  WishlistContext,
  type WishlistContextValue,
} from "@/components/ShopProvider";

/** useWishlist reads the signed-in user's wishlist and the actions that change it. */
export function useWishlist(): WishlistContextValue {
  const wishlist = useContext(WishlistContext);

  if (!wishlist) {
    throw new Error("useWishlist must be used within a ShopProvider");
  }

  return wishlist;
}

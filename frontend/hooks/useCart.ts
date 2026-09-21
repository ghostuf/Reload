"use client";

import { useContext } from "react";
import { CartContext, type CartContextValue } from "@/components/ShopProvider";

/** useCart reads the signed-in user's cart and the actions that change it. */
export function useCart(): CartContextValue {
  const cart = useContext(CartContext);

  if (!cart) {
    throw new Error("useCart must be used within a ShopProvider");
  }

  return cart;
}

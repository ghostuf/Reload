import type { Metadata } from "next";
import CartView from "@/components/CartView";

export const metadata: Metadata = {
  title: "Cart — Reload",
  description: "The games in your Reload cart.",
};

export default function CartPage() {
  return <CartView />;
}

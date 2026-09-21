import type { Metadata } from "next";
import WishlistView from "@/components/WishlistView";

export const metadata: Metadata = {
  title: "Wishlist — Reload",
  description: "The games you have saved on Reload.",
};

export default function WishlistPage() {
  return <WishlistView />;
}

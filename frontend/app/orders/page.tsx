import type { Metadata } from "next";
import OrderHistoryView from "@/components/OrderHistoryView";

export const metadata: Metadata = {
  title: "Order History — Reload",
  description: "The games you have bought on Reload.",
};

export default function OrdersPage() {
  return <OrderHistoryView />;
}

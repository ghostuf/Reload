import type { Metadata } from "next";
import OrderView from "@/components/OrderView";

export const metadata: Metadata = {
  title: "Order — Reload",
  description: "Your Reload order receipt.",
};

export default async function OrderPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return <OrderView orderId={id} />;
}

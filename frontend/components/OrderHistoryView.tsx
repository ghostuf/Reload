"use client";

import Link from "next/link";
import RequireAuth from "@/components/RequireAuth";
import { formatDate } from "@/components/OrderView";
import { useLoad } from "@/hooks/useLoad";
import { getOrders } from "@/lib/orders";
import type { OrderSummary } from "@/types/order";

export default function OrderHistoryView() {
  return (
    <RequireAuth>
      <OrderHistory />
    </RequireAuth>
  );
}

function OrderHistory() {
  // getOrders is a module-level function, so its identity is already stable.
  const { data, loading, error } = useLoad(getOrders);
  const orders = data ?? [];

  return (
    <main className="flex-1">
      <section className="mx-auto max-w-[1440px] px-6 pb-8 pt-16 md:px-12 md:pt-24">
        <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
          Your Purchases
        </p>
        <h1 className="mt-4 text-4xl font-black tracking-tight text-on-surface md:text-5xl">
          Order History
        </h1>
        {!loading && orders.length > 0 && (
          <p className="mt-2 text-sm text-on-surface-variant">
            {orders.length} {orders.length === 1 ? "order" : "orders"}
          </p>
        )}
      </section>

      <section className="mx-auto max-w-[1440px] px-6 pb-20 md:px-12">
        <div className="max-w-3xl">
          {loading ? (
            <OrderHistorySkeleton />
          ) : error ? (
            <p
              role="alert"
              className="rounded-game-card border border-red-400/30 bg-red-400/5 px-6 py-5 text-sm text-red-400"
            >
              {error}
            </p>
          ) : orders.length === 0 ? (
            <EmptyHistory />
          ) : (
            <ul className="flex flex-col gap-3">
              {orders.map((order) => (
                <OrderRow key={order.id} order={order} />
              ))}
            </ul>
          )}
        </div>
      </section>
    </main>
  );
}

function OrderRow({ order }: { order: OrderSummary }) {
  return (
    <li>
      <Link
        href={`/orders/${order.id}`}
        className="group flex flex-wrap items-center justify-between gap-4 rounded-game-card border border-white/5 bg-surface-container-low p-4 transition-colors hover:border-white/20 hover:bg-surface-container sm:p-5"
      >
        <div className="min-w-0">
          <p className="font-bold text-on-surface transition-colors group-hover:text-electric-blue">
            Order #{order.id}
          </p>
          <p className="mt-1 text-sm text-on-surface-variant">
            {formatDate(order.created_at)} · {order.item_count}{" "}
            {order.item_count === 1 ? "game" : "games"}
          </p>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-lg font-black text-on-surface">
            ${order.total_price.toFixed(2)}
          </span>
          <span
            aria-hidden
            className="text-on-surface-variant transition-transform group-hover:translate-x-0.5"
          >
            →
          </span>
        </div>
      </Link>
    </li>
  );
}

function ReceiptIcon({ className = "h-6 w-6" }: { className?: string }) {
  return (
    <svg
      aria-hidden
      viewBox="0 0 24 24"
      className={className}
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M5 3.5v17l2.3-1.6 2.3 1.6 2.4-1.6 2.3 1.6 2.4-1.6 2.3 1.6v-17L16.7 5l-2.3-1.5L12 5 9.7 3.5 7.3 5Z" />
      <path d="M9 9.5h6M9 13.5h4" />
    </svg>
  );
}

function EmptyHistory() {
  return (
    <div className="flex flex-col items-center gap-3 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-24 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-full bg-white/5 text-on-surface-variant">
        <ReceiptIcon />
      </span>
      <p className="text-lg font-semibold text-on-surface">No orders yet</p>
      <p className="text-sm text-on-surface-variant">
        Games you buy will show up here with their receipts.
      </p>
      <Link
        href="/"
        className="mt-3 rounded-xl bg-electric-blue px-5 py-2.5 text-sm font-bold text-white transition-transform active:scale-95"
      >
        Browse Games
      </Link>
    </div>
  );
}

function OrderHistorySkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <ul aria-hidden className="flex flex-col gap-3">
      {Array.from({ length: rows }, (_, i) => (
        <li
          key={i}
          className="flex animate-pulse items-center justify-between gap-4 rounded-game-card border border-white/5 bg-surface-container-low p-4 sm:p-5"
        >
          <div className="flex flex-col gap-2">
            <div className="h-4 w-28 rounded bg-white/5" />
            <div className="h-3 w-40 rounded bg-white/5" />
          </div>
          <div className="h-5 w-16 rounded bg-white/5" />
        </li>
      ))}
    </ul>
  );
}

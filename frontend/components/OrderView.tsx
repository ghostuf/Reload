"use client";

import Image from "next/image";
import Link from "next/link";
import { useCallback } from "react";
import RequireAuth from "@/components/RequireAuth";
import { useLoad } from "@/hooks/useLoad";
import { getOrder } from "@/lib/orders";
import { formatPrice } from "@/lib/pricing";
import type { OrderItem } from "@/types/order";
import ErrorState from "@/components/ErrorState";

export default function OrderView({ orderId }: { orderId: string }) {
  return (
    <RequireAuth>
      <OrderReceipt orderId={orderId} />
    </RequireAuth>
  );
}

/** formatDate renders an order's timestamp the way the account page does. */
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

const STATUS_STYLES: Record<string, string> = {
  completed: "border-neon-lime/40 bg-neon-lime/10 text-neon-lime",
  pending: "border-amber-400/40 bg-amber-400/10 text-amber-400",
  failed: "border-red-400/40 bg-red-400/10 text-red-400",
  cancelled: "border-white/15 bg-white/5 text-on-surface-variant",
};

/**
 * OrderStatusBadge is shared between the receipt and the history list, so
 * "pending", "failed", and "cancelled" -- which only an eSewa-checkout order
 * can ever actually be -- all read consistently wherever an order's status
 * shows up.
 */
export function OrderStatusBadge({ status }: { status: string }) {
  const style = STATUS_STYLES[status] ?? STATUS_STYLES.cancelled;
  return (
    <span
      className={`rounded-full border px-3 py-1 text-[11px] font-bold uppercase tracking-wide ${style}`}
    >
      {status}
    </span>
  );
}

function OrderReceipt({ orderId }: { orderId: string }) {
  const load = useCallback(() => getOrder(orderId), [orderId]);
  const { data: order, loading, error, retry } = useLoad(load);

  const items = order?.items ?? [];

  return (
    <main className="flex-1">
      <section className="mx-auto max-w-[1440px] px-4 sm:px-6 pb-8 pt-16 md:px-12 md:pt-24">
        <Link
          href="/orders"
          className="text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface"
        >
          ← Order History
        </Link>
        <p className="mt-6 text-sm font-semibold uppercase tracking-widest text-electric-blue">
          Order Confirmed
        </p>
        <h1 className="mt-4 text-4xl font-black tracking-tight text-on-surface md:text-5xl">
          Order #{orderId}
        </h1>
        {order && (
          <p className="mt-2 text-sm text-on-surface-variant">
            Placed {formatDate(order.created_at)}
          </p>
        )}
      </section>

      <section className="mx-auto max-w-[1440px] px-4 sm:px-6 pb-20 md:px-12">
        <div className="max-w-3xl">
          {loading ? (
            <ReceiptSkeleton />
          ) : error ? (
            <ErrorState
              title="Couldn't load this order"
              message={error}
              onRetry={retry}
            />
          ) : !order ? (
            <MissingOrder />
          ) : (
            <div className="rounded-[24px] border border-white/10 bg-surface-container-low/80 p-6 backdrop-blur-xl md:p-8">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 pb-5">
                <h2 className="text-lg font-bold text-on-surface">Receipt</h2>
                <OrderStatusBadge status={order.status} />
              </div>

              {order.payment && (
                <p className="mt-3 text-xs text-on-surface-variant">
                  Paid via eSewa
                  {order.payment.provider_ref_id && (
                    <>
                      {" "}
                      · Ref{" "}
                      <span className="font-mono text-on-surface">
                        {order.payment.provider_ref_id}
                      </span>
                    </>
                  )}
                </p>
              )}

              <ul className="flex flex-col divide-y divide-white/5">
                {items.map((item) => (
                  <ReceiptRow key={item.id} item={item} />
                ))}
              </ul>

              <dl className="flex items-baseline justify-between border-t border-white/10 pt-5">
                <dt className="font-semibold text-on-surface">
                  Total · {items.length}{" "}
                  {items.length === 1 ? "item" : "items"}
                </dt>
                <dd className="text-2xl font-black text-on-surface">
                  {formatPrice(order.total_price)}
                </dd>
              </dl>

              <div className="mt-8 flex flex-wrap items-center gap-3">
                <Link
                  href="/library"
                  className="rounded-xl bg-electric-blue px-6 py-3 text-sm font-bold text-white transition-transform active:scale-95"
                >
                  Go to Library
                </Link>
                <Link
                  href="/"
                  className="rounded-xl border border-white/15 px-6 py-3 text-sm font-bold text-on-surface transition-colors hover:border-white/30 hover:bg-surface-container"
                >
                  Keep Browsing
                </Link>
              </div>
            </div>
          )}
        </div>
      </section>
    </main>
  );
}

/**
 * ReceiptRow shows what the game cost at purchase, not what it costs now: a
 * receipt that changed with the store price would stop being a record of what
 * the customer actually paid.
 */
function ReceiptRow({ item }: { item: OrderItem }) {
  return (
    <li className="flex items-center gap-4 py-4">
      <Link
        href={`/games/${item.game_id}`}
        className="relative aspect-[3/4] w-14 shrink-0 overflow-hidden rounded-lg"
      >
        <Image
          src={item.game.cover_image_url}
          alt={item.game.title}
          fill
          sizes="56px"
          className="object-cover"
        />
      </Link>
      <div className="min-w-0 flex-1">
        <Link
          href={`/games/${item.game_id}`}
          className="line-clamp-1 font-bold text-on-surface transition-colors hover:text-electric-blue"
        >
          {item.game.title}
        </Link>
        <p className="mt-0.5 truncate text-sm text-on-surface-variant">
          {item.game.developer}
        </p>
      </div>
      <span className="shrink-0 font-bold text-on-surface">
        {formatPrice(item.price_at_purchase)}
      </span>
    </li>
  );
}

function MissingOrder() {
  return (
    <div className="flex flex-col items-center gap-3 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-24 text-center">
      <p className="text-lg font-semibold text-on-surface">Order not found</p>
      <p className="text-sm text-on-surface-variant">
        This order doesn&apos;t exist on your account.
      </p>
      <Link
        href="/orders"
        className="mt-3 rounded-xl bg-electric-blue px-5 py-2.5 text-sm font-bold text-white transition-transform active:scale-95"
      >
        View Your Orders
      </Link>
    </div>
  );
}

function ReceiptSkeleton() {
  return (
    <div
      aria-hidden
      className="animate-pulse rounded-[24px] border border-white/10 bg-surface-container-low/80 p-6 md:p-8"
    >
      <div className="h-5 w-24 rounded bg-white/5" />
      <div className="mt-6 flex flex-col gap-5">
        {Array.from({ length: 2 }, (_, i) => (
          <div key={i} className="flex items-center gap-4">
            <div className="aspect-[3/4] w-14 shrink-0 rounded-lg bg-white/5" />
            <div className="flex flex-1 flex-col gap-2">
              <div className="h-4 w-1/2 rounded bg-white/5" />
              <div className="h-3 w-1/3 rounded bg-white/5" />
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 h-8 w-32 self-end rounded bg-white/5" />
    </div>
  );
}

import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { CartButton, WishlistButton } from "@/components/GameActions";
import { getGame } from "@/lib/api";

export default async function GameDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const game = await getGame(id);

  if (!game) {
    notFound();
  }

  const hasDiscount =
    game.discount_price != null && game.discount_price < game.price;
  const releaseDate = new Date(game.release_date);

  return (
    <main className="mx-auto max-w-[1440px] flex-1 px-6 py-12 md:px-12">
      <Link
        href="/"
        className="text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface"
      >
        ← Back to Store
      </Link>

      <div className="mt-6 grid gap-10 lg:grid-cols-[380px_1fr]">
        <div className="relative aspect-[3/4] overflow-hidden rounded-game-card">
          <Image
            src={game.cover_image_url}
            alt={game.title}
            fill
            sizes="380px"
            className="object-cover"
            preload
          />
        </div>

        <div className="flex flex-col gap-6">
          <div>
            <div className="flex flex-wrap gap-2">
              {game.genres?.map((genre) => (
                <span
                  key={genre.id}
                  className="rounded-full bg-white/5 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-on-surface-variant"
                >
                  {genre.name}
                </span>
              ))}
            </div>
            <h1 className="mt-4 text-3xl font-black tracking-tight text-on-surface md:text-5xl">
              {game.title}
            </h1>
            <p className="mt-2 text-sm text-on-surface-variant">
              {game.developer} · {game.publisher} ·{" "}
              {releaseDate.toLocaleDateString(undefined, {
                year: "numeric",
                month: "long",
                day: "numeric",
              })}
            </p>
          </div>

          <div className="flex items-center gap-3">
            {hasDiscount ? (
              <>
                <span className="text-lg text-on-surface-variant line-through">
                  ${game.price.toFixed(2)}
                </span>
                <span className="text-2xl font-bold text-on-surface">
                  ${game.discount_price!.toFixed(2)}
                </span>
              </>
            ) : (
              <span className="text-2xl font-bold text-on-surface">
                ${game.price.toFixed(2)}
              </span>
            )}
          </div>

          <div className="flex flex-wrap items-start gap-3">
            <CartButton gameId={game.id} variant="detail" />
            <WishlistButton gameId={game.id} variant="detail" />
          </div>

          <p className="max-w-2xl leading-relaxed text-on-surface-variant">
            {game.description}
          </p>
        </div>
      </div>

      {game.images && game.images.length > 0 && (
        <div className="mt-12">
          <h2 className="text-xl font-bold text-on-surface">Gallery</h2>
          <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {game.images.map((image) => (
              <div
                key={image.id}
                className="relative aspect-video overflow-hidden rounded-game-card"
              >
                <Image
                  src={image.image_url}
                  alt={`${game.title} screenshot`}
                  fill
                  sizes="(max-width: 768px) 100vw, 33vw"
                  className="object-cover"
                />
              </div>
            ))}
          </div>
        </div>
      )}
    </main>
  );
}

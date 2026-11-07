import Link from "next/link";

export default function GameNotFound() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4 px-4 py-32 text-center sm:px-6">
      <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
        404
      </p>
      <h1 className="text-3xl font-black tracking-tight text-on-surface md:text-5xl">
        Game not found.
      </h1>
      <p className="max-w-md text-on-surface-variant">
        This game isn&apos;t in the store. It may have been removed, or the link
        may be wrong.
      </p>
      <Link
        href="/"
        className="mt-4 rounded-xl bg-electric-blue px-6 py-3 text-sm font-bold text-white transition-transform active:scale-95"
      >
        Browse the store
      </Link>
    </main>
  );
}

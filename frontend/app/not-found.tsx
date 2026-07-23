import Link from "next/link";

export default function NotFound() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4 px-6 py-32 text-center">
      <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
        404
      </p>
      <h1 className="text-3xl font-black tracking-tight text-on-surface md:text-5xl">
        This page went off the grid.
      </h1>
      <p className="max-w-md text-on-surface-variant">
        The game or page you&apos;re looking for doesn&apos;t exist.
      </p>
      <Link
        href="/"
        className="mt-4 rounded-xl bg-electric-blue px-6 py-3 text-sm font-bold text-white transition-transform active:scale-95"
      >
        Back to Store
      </Link>
    </main>
  );
}

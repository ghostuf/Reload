"use client";

export default function Error({ reset }: { error: Error; reset: () => void }) {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4 px-6 py-32 text-center">
      <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
        Error
      </p>
      <h1 className="text-3xl font-black tracking-tight text-on-surface md:text-5xl">
        Something broke.
      </h1>
      <p className="max-w-md text-on-surface-variant">
        An unexpected error occurred while loading this page.
      </p>
      <button
        onClick={reset}
        className="mt-4 rounded-xl bg-electric-blue px-6 py-3 text-sm font-bold text-white transition-transform active:scale-95"
      >
        Try again
      </button>
    </main>
  );
}

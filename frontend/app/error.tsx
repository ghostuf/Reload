"use client";

/**
 * The fallback for anything that throws while rendering — most often the API
 * being unreachable. unstable_retry re-fetches the failed segment (reset would
 * only re-render it with the same failed data), so "Try again" can recover
 * once the API is back.
 */
export default function Error({
  unstable_retry,
}: {
  error: Error & { digest?: string };
  unstable_retry: () => void;
}) {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4 px-4 py-32 text-center sm:px-6">
      <p className="text-sm font-semibold uppercase tracking-widest text-electric-blue">
        Error
      </p>
      <h1 className="text-3xl font-black tracking-tight text-on-surface md:text-5xl">
        Something broke.
      </h1>
      <p className="max-w-md text-on-surface-variant">
        We couldn&apos;t load this page. The store may be briefly unavailable.
      </p>
      <button
        type="button"
        onClick={() => unstable_retry()}
        className="mt-4 rounded-xl bg-electric-blue px-6 py-3 text-sm font-bold text-white transition-transform active:scale-95"
      >
        Try again
      </button>
    </main>
  );
}

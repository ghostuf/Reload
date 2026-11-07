"use client";

/**
 * ErrorState is the friendly "that didn't load" panel: what failed, the API's
 * own message, and a way to try again.
 */
export default function ErrorState({
  title = "Something went wrong",
  message,
  onRetry,
}: {
  title?: string;
  message?: string | null;
  onRetry?: () => void;
}) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center gap-2 rounded-game-card border border-white/5 bg-surface-container-low px-6 py-16 text-center"
    >
      <span
        aria-hidden
        className="flex h-12 w-12 items-center justify-center rounded-full bg-red-400/10 text-xl font-black text-red-400"
      >
        !
      </span>
      <p className="mt-2 text-lg font-semibold text-on-surface">{title}</p>
      {message && <p className="max-w-md text-sm text-on-surface-variant">{message}</p>}
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-3 rounded-xl bg-electric-blue px-5 py-2.5 text-sm font-bold text-white transition-transform active:scale-95"
        >
          Try again
        </button>
      )}
    </div>
  );
}

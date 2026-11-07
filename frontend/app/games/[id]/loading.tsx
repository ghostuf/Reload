/** Shown while a game's page is fetched; mirrors the detail layout. */
export default function GameDetailLoading() {
  return (
    <main
      aria-busy
      aria-label="Loading game"
      className="mx-auto w-full max-w-[1440px] flex-1 px-4 py-12 sm:px-6 md:px-12"
    >
      <div className="h-4 w-28 animate-pulse rounded bg-white/5" />
      <div className="mt-6 grid animate-pulse gap-8 lg:grid-cols-[380px_1fr] lg:gap-10">
        <div className="mx-auto aspect-[3/4] w-full max-w-sm rounded-game-card bg-surface-container-low lg:max-w-none" />
        <div className="flex flex-col gap-6">
          <div>
            <div className="flex gap-2">
              <div className="h-6 w-16 rounded-full bg-white/5" />
              <div className="h-6 w-20 rounded-full bg-white/5" />
            </div>
            <div className="mt-4 h-10 w-3/4 rounded bg-white/5 md:h-14" />
            <div className="mt-3 h-4 w-1/2 rounded bg-white/5" />
            <div className="mt-4 h-4 w-40 rounded bg-white/5" />
          </div>
          <div className="h-8 w-32 rounded bg-white/5" />
          <div className="flex gap-3">
            <div className="h-12 w-40 rounded-xl bg-white/5" />
            <div className="h-12 w-12 rounded-xl bg-white/5" />
          </div>
          <div className="flex max-w-2xl flex-col gap-2">
            <div className="h-4 w-full rounded bg-white/5" />
            <div className="h-4 w-full rounded bg-white/5" />
            <div className="h-4 w-2/3 rounded bg-white/5" />
          </div>
        </div>
      </div>
    </main>
  );
}

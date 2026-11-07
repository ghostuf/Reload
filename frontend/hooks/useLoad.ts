"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

export type Loadable<T> = {
  /** The loaded value, or null while loading and after a failure. */
  data: T | null;
  loading: boolean;
  /** Set when the load failed; `data` is then null. */
  error: string | null;
  /** Runs the read again, showing the loading state until it settles. */
  retry: () => void;
};

type Result<T> = {
  /** The read this result came from. */
  source: () => Promise<T>;
  /** Which retry of that read it came from. */
  attempt: number;
  data: T | null;
  error: string | null;
};

/**
 * useLoad runs a one-shot read and reports its three states.
 *
 * Orders and receipts live behind the session cookies, which only the browser
 * holds, so they are fetched from the client rather than rendered on the
 * server. `load` must be stable — a module-level function, or one wrapped in
 * useCallback — since a new identity re-runs the read.
 *
 * The result is tagged with the read it came from rather than cleared when
 * `load` changes: anything tagged with a superseded read simply reads as
 * loading, which keeps a stale value from showing for a frame without a reset
 * step that could race the request in flight.
 */
export function useLoad<T>(load: () => Promise<T>): Loadable<T> {
  const [result, setResult] = useState<Result<T> | null>(null);
  const [attempt, setAttempt] = useState(0);

  const current =
    result !== null && result.source === load && result.attempt === attempt
      ? result
      : null;

  useEffect(() => {
    let active = true;

    load()
      .then((data) => {
        if (!active) return;
        setResult({ source: load, attempt, data, error: null });
      })
      .catch((err: unknown) => {
        if (!active) return;
        setResult({
          source: load,
          attempt,
          data: null,
          error:
            err instanceof Error && err.message
              ? err.message
              : "Something went wrong",
        });
      });

    return () => {
      // A response that lands after `load` changed belongs to a read the
      // caller is no longer asking for.
      active = false;
    };
  }, [load, attempt]);

  const retry = useCallback(() => setAttempt((a) => a + 1), []);

  return useMemo(
    () => ({
      data: current?.data ?? null,
      loading: current === null,
      error: current?.error ?? null,
      retry,
    }),
    [current, retry],
  );
}

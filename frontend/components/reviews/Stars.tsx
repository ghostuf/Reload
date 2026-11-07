"use client";

import { useState } from "react";

const STAR_PATH =
  "M12 2.5l2.9 6.1 6.6.8-4.9 4.6 1.3 6.6L12 17.3l-5.9 3.3 1.3-6.6-4.9-4.6 6.6-.8z";

function Star({ fill, size }: { fill: number; size: number }) {
  // fill runs 0..1; the clip shows that fraction of a lit star over a dim one.
  const width = `${Math.max(0, Math.min(1, fill)) * 100}%`;
  return (
    <span className="relative inline-block" style={{ width: size, height: size }}>
      <svg viewBox="0 0 24 24" width={size} height={size} className="text-white/15" aria-hidden>
        <path d={STAR_PATH} fill="currentColor" />
      </svg>
      <span className="absolute inset-0 overflow-hidden" style={{ width }}>
        <svg viewBox="0 0 24 24" width={size} height={size} className="text-neon-lime" aria-hidden>
          <path d={STAR_PATH} fill="currentColor" />
        </svg>
      </span>
    </span>
  );
}

/** Stars shows a rating out of five, including fractional averages. */
export function Stars({ rating, size = 16 }: { rating: number; size?: number }) {
  return (
    <span
      role="img"
      aria-label={`${rating.toFixed(1)} out of 5 stars`}
      className="inline-flex items-center gap-0.5"
    >
      {[0, 1, 2, 3, 4].map((i) => (
        <Star key={i} fill={rating - i} size={size} />
      ))}
    </span>
  );
}

/**
 * StarInput picks a whole-star rating. It is a radio group under the hood, so
 * it works with the keyboard (arrow keys) and reads correctly to screen readers.
 */
export function StarInput({
  value,
  onChange,
  disabled = false,
}: {
  value: number;
  onChange: (rating: number) => void;
  disabled?: boolean;
}) {
  const [hover, setHover] = useState(0);
  const shown = hover || value;

  return (
    <div
      role="radiogroup"
      aria-label="Rating"
      className="inline-flex items-center gap-1"
      onMouseLeave={() => setHover(0)}
    >
      {[1, 2, 3, 4, 5].map((n) => (
        <label
          key={n}
          className={`cursor-pointer rounded p-0.5 focus-within:outline focus-within:outline-2 focus-within:outline-electric-blue ${
            disabled ? "cursor-not-allowed opacity-60" : ""
          }`}
          onMouseEnter={() => !disabled && setHover(n)}
        >
          <input
            type="radio"
            name="rating"
            value={n}
            checked={value === n}
            onChange={() => onChange(n)}
            disabled={disabled}
            className="sr-only"
            aria-label={`${n} ${n === 1 ? "star" : "stars"}`}
          />
          <Star fill={shown >= n ? 1 : 0} size={26} />
        </label>
      ))}
    </div>
  );
}

import Link from "next/link";

// Iteration 1 has no GET /genres endpoint, so the taxonomy is hardcoded to
// match the seeded data. Revisit if a genres endpoint is added later.
const GENRES = [
  "Action",
  "RPG",
  "Adventure",
  "Indie",
  "Multiplayer",
  "Horror",
  "Racing",
  "Strategy",
  "Simulation",
  "Sports",
];

type Props = {
  active?: string;
  searchParams: Record<string, string | undefined>;
};

export default function GenreChips({ active, searchParams }: Props) {
  function hrefFor(genre?: string) {
    const params = new URLSearchParams();
    for (const [key, value] of Object.entries(searchParams)) {
      if (value && key !== "genre" && key !== "page") params.set(key, value);
    }
    if (genre) params.set("genre", genre);
    const query = params.toString();
    return query ? `/?${query}` : "/";
  }

  return (
    <div className="flex flex-wrap gap-2">
      <Link href={hrefFor(undefined)} className={chipClass(!active)}>
        All
      </Link>
      {GENRES.map((genre) => (
        <Link
          key={genre}
          href={hrefFor(genre)}
          className={chipClass(active?.toLowerCase() === genre.toLowerCase())}
        >
          {genre}
        </Link>
      ))}
    </div>
  );
}

function chipClass(isActive: boolean) {
  return `rounded-full border px-4 py-2 text-sm font-semibold transition-colors ${
    isActive
      ? "border-electric-blue bg-electric-blue/20 text-on-surface"
      : "border-white/10 bg-white/5 text-on-surface-variant hover:bg-white/10"
  }`;
}

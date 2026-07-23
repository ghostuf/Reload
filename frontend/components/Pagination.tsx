import Link from "next/link";

type Props = {
  page: number;
  totalPages: number;
  searchParams: Record<string, string | undefined>;
};

function buildHref(
  searchParams: Record<string, string | undefined>,
  page: number,
) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(searchParams)) {
    if (value) params.set(key, value);
  }
  params.set("page", String(page));
  return `/?${params.toString()}`;
}

export default function Pagination({ page, totalPages, searchParams }: Props) {
  if (totalPages <= 1) return null;

  const prevDisabled = page <= 1;
  const nextDisabled = page >= totalPages;

  return (
    <div className="flex items-center justify-center gap-4">
      <Link
        href={buildHref(searchParams, page - 1)}
        aria-disabled={prevDisabled}
        className={`rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold transition-colors ${
          prevDisabled
            ? "pointer-events-none opacity-30"
            : "text-on-surface hover:bg-white/10"
        }`}
      >
        Previous
      </Link>
      <span className="text-sm text-on-surface-variant">
        Page {page} of {totalPages}
      </span>
      <Link
        href={buildHref(searchParams, page + 1)}
        aria-disabled={nextDisabled}
        className={`rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold transition-colors ${
          nextDisabled
            ? "pointer-events-none opacity-30"
            : "text-on-surface hover:bg-white/10"
        }`}
      >
        Next
      </Link>
    </div>
  );
}

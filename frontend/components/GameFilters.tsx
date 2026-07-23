"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";

const SORT_OPTIONS = [
  { value: "newest", label: "Newest" },
  { value: "price_asc", label: "Price: Low to High" },
  { value: "price_desc", label: "Price: High to Low" },
  { value: "title_asc", label: "Title: A-Z" },
  { value: "title_desc", label: "Title: Z-A" },
  { value: "release_date", label: "Release Date" },
];

export default function GameFilters() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [search, setSearch] = useState(searchParams.get("search") ?? "");

  function updateParams(next: Record<string, string | null>) {
    const params = new URLSearchParams(searchParams.toString());
    for (const [key, value] of Object.entries(next)) {
      if (value) params.set(key, value);
      else params.delete(key);
    }
    params.delete("page");
    router.push(`/?${params.toString()}`);
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    updateParams({ search: search.trim() || null });
  }

  return (
    <div className="flex flex-col gap-4 rounded-[24px] border border-white/10 bg-surface-container-low/80 p-4 backdrop-blur-xl md:flex-row md:items-center md:p-6">
      <form
        onSubmit={handleSubmit}
        className="flex flex-1 items-center gap-3 rounded-2xl border border-white/10 bg-surface-container-lowest px-5 py-3 focus-within:border-electric-blue"
      >
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search for games..."
          className="w-full bg-transparent text-on-surface outline-none placeholder:text-on-surface-variant"
        />
        <button
          type="submit"
          className="shrink-0 rounded-xl bg-electric-blue px-4 py-1.5 text-sm font-bold text-white transition-transform active:scale-95"
        >
          Search
        </button>
      </form>
      <select
        value={searchParams.get("sort") ?? "newest"}
        onChange={(e) => updateParams({ sort: e.target.value })}
        className="rounded-2xl border border-white/10 bg-surface-container-lowest px-4 py-3 text-sm text-on-surface outline-none focus:border-electric-blue"
      >
        {SORT_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}

"use client";

import Link from "next/link";
import { useState } from "react";
import type { AdminGameInput } from "@/lib/admin";
import { GENRES } from "@/lib/genres";
import {
  EMPTY_GAME_FORM,
  slugify,
  validateGameForm,
  type GameFormErrors,
  type GameFormValues,
} from "@/lib/gameForm";

/** Hosts the store's next/image is configured for; see next.config.ts. */
const STORE_IMAGE_HOSTS = ["picsum.photos"];

function hostOf(url: string): string | null {
  try {
    return new URL(url).hostname;
  } catch {
    return null;
  }
}

const inputClass =
  "w-full rounded-2xl border bg-surface-container-lowest px-4 py-2.5 text-on-surface outline-none placeholder:text-on-surface-variant focus:border-electric-blue";

function borderFor(error?: string) {
  return error ? "border-red-400/60" : "border-white/10";
}

/**
 * GameForm creates or edits a game. New games get a slug from the title until
 * the slug is edited by hand. Client checks catch the obvious mistakes; the
 * API's own rejections (a taken slug, say) show at the bottom.
 */
export default function GameForm({
  initial = EMPTY_GAME_FORM,
  isNew,
  onSubmit,
}: {
  initial?: GameFormValues;
  isNew: boolean;
  /** Throws to report a failure; its message is shown under the form. */
  onSubmit: (input: AdminGameInput) => Promise<void>;
}) {
  const [values, setValues] = useState<GameFormValues>(initial);
  const [slugTouched, setSlugTouched] = useState(!isNew);
  const [errors, setErrors] = useState<GameFormErrors>({});
  const [serverError, setServerError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [customGenre, setCustomGenre] = useState("");

  function set<K extends keyof GameFormValues>(key: K, value: GameFormValues[K]) {
    setValues((prev) => {
      const next = { ...prev, [key]: value };
      if (key === "title" && !slugTouched) next.slug = slugify(String(value));
      return next;
    });
  }

  function toggleGenre(name: string) {
    const has = values.genres.some((g) => g.toLowerCase() === name.toLowerCase());
    set(
      "genres",
      has
        ? values.genres.filter((g) => g.toLowerCase() !== name.toLowerCase())
        : [...values.genres, name],
    );
  }

  function addCustomGenre() {
    const name = customGenre.trim();
    if (!name) return;
    if (!values.genres.some((g) => g.toLowerCase() === name.toLowerCase())) {
      set("genres", [...values.genres, name]);
    }
    setCustomGenre("");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const result = validateGameForm(values);
    if (!result.ok) {
      setErrors(result.errors);
      setServerError(null);
      return;
    }

    setErrors({});
    setServerError(null);
    setSubmitting(true);
    try {
      await onSubmit(result.input);
    } catch (err) {
      setServerError(err instanceof Error && err.message ? err.message : "Could not save the game");
      setSubmitting(false);
    }
    // On success the caller navigates away, so submitting stays true.
  }

  // Every genre on offer: the store's list plus anything this game already has.
  const genreOptions = [
    ...GENRES,
    ...values.genres.filter((g) => !GENRES.some((s) => s.toLowerCase() === g.toLowerCase())),
  ];
  const coverHost = hostOf(values.coverImageUrl.trim());
  const coverHostUnsupported = coverHost !== null && !STORE_IMAGE_HOSTS.includes(coverHost);
  const errorCount = Object.keys(errors).length;

  return (
    <form onSubmit={handleSubmit} noValidate aria-label="Game form" className="mt-6 flex max-w-3xl flex-col gap-8">
      <Section title="Details">
        <Field label="Title" htmlFor="title" required error={errors.title}>
          <input id="title" value={values.title} onChange={(e) => set("title", e.target.value)} className={`${inputClass} ${borderFor(errors.title)}`} />
        </Field>
        <Field label="Slug" htmlFor="slug" required error={errors.slug} hint="Used in URLs. Must be unique.">
          <input
            id="slug"
            value={values.slug}
            onChange={(e) => {
              setSlugTouched(true);
              set("slug", e.target.value);
            }}
            className={`${inputClass} ${borderFor(errors.slug)}`}
          />
        </Field>
        <Field label="Description" htmlFor="description">
          <textarea id="description" rows={5} value={values.description} onChange={(e) => set("description", e.target.value)} className={`${inputClass} ${borderFor()} resize-y`} />
        </Field>
        <div className="grid gap-5 sm:grid-cols-2">
          <Field label="Developer" htmlFor="developer">
            <input id="developer" value={values.developer} onChange={(e) => set("developer", e.target.value)} className={`${inputClass} ${borderFor()}`} />
          </Field>
          <Field label="Publisher" htmlFor="publisher">
            <input id="publisher" value={values.publisher} onChange={(e) => set("publisher", e.target.value)} className={`${inputClass} ${borderFor()}`} />
          </Field>
        </div>
        <Field label="Release date" htmlFor="releaseDate" error={errors.releaseDate}>
          <input id="releaseDate" type="date" value={values.releaseDate} onChange={(e) => set("releaseDate", e.target.value)} className={`${inputClass} ${borderFor(errors.releaseDate)} [color-scheme:dark] sm:max-w-xs`} />
        </Field>
      </Section>

      <Section title="Pricing">
        <div className="grid gap-5 sm:grid-cols-2">
          <Field label="Price (USD)" htmlFor="price" required error={errors.price}>
            <input id="price" inputMode="decimal" placeholder="29.99" value={values.price} onChange={(e) => set("price", e.target.value)} className={`${inputClass} ${borderFor(errors.price)}`} />
          </Field>
          <Field label="Discount price" htmlFor="discountPrice" error={errors.discountPrice} hint="Leave empty for no discount.">
            <input id="discountPrice" inputMode="decimal" placeholder="19.99" value={values.discountPrice} onChange={(e) => set("discountPrice", e.target.value)} className={`${inputClass} ${borderFor(errors.discountPrice)}`} />
          </Field>
        </div>
        <Field label="Discount ends" htmlFor="discountEndsAt" error={errors.discountEndsAt} hint="Your local time. Leave empty for a discount with no end date.">
          <input id="discountEndsAt" type="datetime-local" value={values.discountEndsAt} onChange={(e) => set("discountEndsAt", e.target.value)} className={`${inputClass} ${borderFor(errors.discountEndsAt)} [color-scheme:dark] sm:max-w-xs`} />
        </Field>
      </Section>

      <Section title="Genres">
        <div className="flex flex-wrap gap-2" role="group" aria-label="Genres">
          {genreOptions.map((genre) => {
            const active = values.genres.some((g) => g.toLowerCase() === genre.toLowerCase());
            return (
              <button
                key={genre}
                type="button"
                aria-pressed={active}
                onClick={() => toggleGenre(genre)}
                className={`rounded-full border px-4 py-1.5 text-sm font-semibold transition-colors ${
                  active
                    ? "border-electric-blue bg-electric-blue/20 text-on-surface"
                    : "border-white/10 bg-white/5 text-on-surface-variant hover:bg-white/10"
                }`}
              >
                {genre}
              </button>
            );
          })}
        </div>
        <div className="flex max-w-sm gap-2">
          <label htmlFor="customGenre" className="sr-only">New genre</label>
          <input
            id="customGenre"
            value={customGenre}
            placeholder="Add another genre"
            onChange={(e) => setCustomGenre(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                addCustomGenre();
              }
            }}
            className={`${inputClass} ${borderFor()}`}
          />
          <button type="button" onClick={addCustomGenre} className="shrink-0 rounded-xl border border-white/10 px-4 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10">
            Add
          </button>
        </div>
      </Section>

      <Section title="Images">
        <Field
          label="Cover image URL"
          htmlFor="coverImageUrl"
          required
          error={errors.coverImageUrl}
          hint={coverHostUnsupported ? undefined : "Shown on store cards and the game page."}
        >
          <input id="coverImageUrl" type="url" placeholder="https://picsum.photos/seed/…/600/800" value={values.coverImageUrl} onChange={(e) => set("coverImageUrl", e.target.value)} className={`${inputClass} ${borderFor(errors.coverImageUrl)}`} />
        </Field>
        {coverHostUnsupported && !errors.coverImageUrl && (
          <p className="-mt-3 text-sm text-amber-300">
            The store only loads images from {STORE_IMAGE_HOSTS.join(", ")}. A cover from {coverHost} will
            not show on store pages until that host is added to next.config.ts.
          </p>
        )}

        <fieldset className="flex flex-col gap-3">
          <legend className="mb-2 text-sm font-semibold text-on-surface">Gallery image URLs</legend>
          {values.images.length === 0 && (
            <p className="text-sm text-on-surface-variant">No gallery images.</p>
          )}
          {values.images.map((url, i) => (
            <div key={i} className="flex gap-2">
              <label htmlFor={`image-${i}`} className="sr-only">Gallery image {i + 1}</label>
              <input
                id={`image-${i}`}
                type="url"
                value={url}
                placeholder="https://…"
                onChange={(e) => set("images", values.images.map((u, j) => (j === i ? e.target.value : u)))}
                className={`${inputClass} ${borderFor(errors.images)}`}
              />
              <button
                type="button"
                onClick={() => set("images", values.images.filter((_, j) => j !== i))}
                aria-label={`Remove gallery image ${i + 1}`}
                className="shrink-0 rounded-xl border border-white/10 px-3 text-sm font-semibold text-on-surface-variant transition-colors hover:bg-white/10 hover:text-red-400"
              >
                Remove
              </button>
            </div>
          ))}
          {errors.images && <p className="text-sm text-red-400">{errors.images}</p>}
          <button
            type="button"
            onClick={() => set("images", [...values.images, ""])}
            className="self-start rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
          >
            Add image
          </button>
        </fieldset>
      </Section>

      {(serverError || errorCount > 0) && (
        <p role="alert" className="rounded-xl border border-red-400/30 bg-red-400/10 px-4 py-3 text-sm text-red-400">
          {serverError ?? `Fix the ${errorCount === 1 ? "highlighted field" : `${errorCount} highlighted fields`} and try again.`}
        </p>
      )}

      <div className="flex flex-wrap gap-3">
        <button
          type="submit"
          disabled={submitting}
          className="rounded-xl bg-electric-blue px-6 py-2.5 text-sm font-bold text-white transition-transform active:scale-95 disabled:opacity-50"
        >
          {submitting ? "Saving…" : isNew ? "Create game" : "Save changes"}
        </button>
        <Link
          href="/admin/games"
          className="rounded-xl border border-white/10 px-6 py-2.5 text-sm font-semibold text-on-surface transition-colors hover:bg-white/10"
        >
          Cancel
        </Link>
      </div>
    </form>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-5 rounded-game-card border border-white/5 bg-surface-container-low p-5 sm:p-6">
      <h2 className="text-lg font-bold text-on-surface">{title}</h2>
      {children}
    </section>
  );
}

function Field({
  label,
  htmlFor,
  required = false,
  error,
  hint,
  children,
}: {
  label: string;
  htmlFor: string;
  required?: boolean;
  error?: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={htmlFor} className="text-sm font-semibold text-on-surface">
        {label}
        {required && <span className="text-red-400" aria-hidden> *</span>}
      </label>
      {children}
      {error ? (
        <p className="text-sm text-red-400">{error}</p>
      ) : hint ? (
        <p className="text-xs text-on-surface-variant">{hint}</p>
      ) : null}
    </div>
  );
}

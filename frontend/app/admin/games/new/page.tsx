"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import GameForm from "@/components/admin/GameForm";
import { createGame } from "@/lib/admin";

export default function NewGamePage() {
  const router = useRouter();

  return (
    <div className="mt-2">
      <Link
        href="/admin/games"
        className="text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface"
      >
        ← All games
      </Link>
      <h1 className="mt-3 text-3xl font-black tracking-tight text-on-surface">New game</h1>
      <GameForm
        isNew
        onSubmit={async (input) => {
          await createGame(input);
          router.push("/admin/games");
        }}
      />
    </div>
  );
}

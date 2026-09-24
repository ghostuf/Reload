import type { Metadata } from "next";
import LibraryView from "@/components/LibraryView";

export const metadata: Metadata = {
  title: "Library — Reload",
  description: "The games you own on Reload.",
};

export default function LibraryPage() {
  return <LibraryView />;
}

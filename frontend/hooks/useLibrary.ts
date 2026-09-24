"use client";

import { useContext } from "react";
import {
  LibraryContext,
  type LibraryContextValue,
} from "@/components/ShopProvider";

/** useLibrary reads the games the signed-in user owns. */
export function useLibrary(): LibraryContextValue {
  const library = useContext(LibraryContext);

  if (!library) {
    throw new Error("useLibrary must be used within a ShopProvider");
  }

  return library;
}

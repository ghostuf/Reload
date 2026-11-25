import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Produces a self-contained .next/standalone build with just the files and
  // pruned node_modules the server needs, so the Docker image doesn't also
  // have to carry the full node_modules tree.
  output: "standalone",
  images: {
    remotePatterns: [
      new URL("https://picsum.photos/**"),
      // Seed data's mock covers and gallery shots -- colored placeholder
      // cards carrying each game's own title, rather than picsum's
      // unrelated stock photography. The object form is required here,
      // not the new URL(...) shorthand: that shorthand sets `search` to
      // the literal (empty) query string of the pattern itself, which
      // Next then reads as "must have no query string" -- rejecting every
      // one of these URLs, which all carry ?text=&font=. Omitting
      // `search` here instead leaves it as the implied wildcard.
      {
        protocol: "https",
        hostname: "placehold.co",
        pathname: "/**",
      },
      // Seed data's covers and gallery shots are now real stock photography
      // served from Unsplash's CDN rather than generated placeholder cards.
      {
        protocol: "https",
        hostname: "images.unsplash.com",
        pathname: "/**",
      },
    ],
  },
};

export default nextConfig;

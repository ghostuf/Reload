import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Produces a self-contained .next/standalone build with just the files and
  // pruned node_modules the server needs, so the Docker image doesn't also
  // have to carry the full node_modules tree.
  output: "standalone",
  images: {
    remotePatterns: [new URL("https://picsum.photos/**")],
  },
};

export default nextConfig;

import type { NextConfig } from "next";
import { createMDX } from "fumadocs-mdx/next";

const nextConfig: NextConfig = {
  // Both packages ship TypeScript source rather than a build: one mark and one set of
  // client configuration shapes, shared with the app so neither can drift between them.
  transpilePackages: ["@nooks/design", "@nooks/shared"],
  /* config options here */
};

const withMDX = createMDX();

export default withMDX(nextConfig);

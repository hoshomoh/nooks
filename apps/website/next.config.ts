import type { NextConfig } from "next";
import { createMDX } from "fumadocs-mdx/next";

const nextConfig: NextConfig = {
  // The design package ships TypeScript source rather than a build: one mark, shared
  // between the app and the site so its geometry cannot drift between them.
  transpilePackages: ["@nooks/design"],
  /* config options here */
};

const withMDX = createMDX();

export default withMDX(nextConfig);

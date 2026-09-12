import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // The design package ships TypeScript source rather than a build: one mark, shared
  // between the app and the site so its geometry cannot drift between them.
  transpilePackages: ["@nooks/design"],
  /* config options here */
};

export default nextConfig;

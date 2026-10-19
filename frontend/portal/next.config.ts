import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  async rewrites() {
    return [
      { source: "/api/core/:path*", destination: `${process.env.NEXT_PUBLIC_CORE_URL || "http://core:4001"}/v1/:path*` },
    ];
  },
};
export default nextConfig;

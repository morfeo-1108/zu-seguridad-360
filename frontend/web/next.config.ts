import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  async rewrites() {
    return [
      { source: "/api/core/:path*", destination: `${process.env.NEXT_PUBLIC_CORE_URL || "http://core:4001"}/v1/:path*` },
      { source: "/api/ai/:path*", destination: `${process.env.NEXT_PUBLIC_AI_URL || "http://ai:4002"}/v1/:path*` },
      { source: "/api/docs/:path*", destination: `${process.env.NEXT_PUBLIC_DOCS_URL || "http://documents:4003"}/v1/:path*` },
    ];
  },
};
export default nextConfig;

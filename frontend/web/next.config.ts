import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  async rewrites() {
    return [
      { source: "/api/core/:path*", destination: `${process.env.NEXT_PUBLIC_CORE_URL || "http://core:4001"}/v1/:path*` },
      { source: "/api/ai/:path*", destination: `${process.env.NEXT_PUBLIC_AI_URL || "http://ai:4002"}/v1/:path*` },
      { source: "/api/docs/:path*", destination: `${process.env.NEXT_PUBLIC_DOCS_URL || "http://documents:4003"}/v1/:path*` },
      { source: "/api/attention/:path*", destination: `${process.env.NEXT_PUBLIC_ATTENTION_URL || "http://attention:4008"}/v1/:path*` },
      { source: "/api/media/:path*", destination: `${process.env.NEXT_PUBLIC_MEDIA_URL || "http://media:4004"}/v1/:path*` },
      { source: "/api/billing/:path*", destination: `${process.env.NEXT_PUBLIC_BILLING_URL || "http://billing:4010"}/v1/:path*` },
    ];
  },
};
export default nextConfig;

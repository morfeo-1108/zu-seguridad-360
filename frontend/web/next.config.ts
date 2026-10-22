import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  // Build into another directory (NEXT_DIST_DIR) while the live server keeps
  // serving .next; scripts/deploy-frontend.sh swaps them. Unset = .next.
  distDir: process.env.NEXT_DIST_DIR || ".next",
  // Behind the Cloudflare Tunnel every request reaches us as plain HTTP, so
  // the original scheme only survives in X-Forwarded-Proto. Send http:// to
  // https:// — Core only accepts the HTTPS origin, so Google sign-in from an
  // http:// page would fail. Local requests (no header) are untouched.
  async redirects() {
    return [
      {
        source: "/:path*",
        has: [
          { type: "header", key: "x-forwarded-proto", value: "http" },
          { type: "host", value: "www.zuseguridad360.com" },
        ],
        destination: "https://www.zuseguridad360.com/:path*",
        permanent: true,
      },
      // The bare domain goes through the same tunnel; canonicalize it to www
      // (the only origin Core and the Google OAuth client allow).
      {
        source: "/:path*",
        has: [{ type: "host", value: "zuseguridad360.com" }],
        destination: "https://www.zuseguridad360.com/:path*",
        permanent: true,
      },
    ];
  },
  async rewrites() {
    return [
      { source: "/api/core/:path*", destination: `${process.env.NEXT_PUBLIC_CORE_URL || "http://core:4001"}/v1/:path*` },
      { source: "/api/ai/:path*", destination: `${process.env.NEXT_PUBLIC_AI_URL || "http://ai:4002"}/v1/:path*` },
      { source: "/api/docs/:path*", destination: `${process.env.NEXT_PUBLIC_DOCS_URL || "http://documents:4003"}/v1/:path*` },
      { source: "/api/attention/:path*", destination: `${process.env.NEXT_PUBLIC_ATTENTION_URL || "http://attention:4008"}/v1/:path*` },
      { source: "/api/media/:path*", destination: `${process.env.NEXT_PUBLIC_MEDIA_URL || "http://media:4004"}/v1/:path*` },
      { source: "/api/billing/:path*", destination: `${process.env.NEXT_PUBLIC_BILLING_URL || "http://billing:4010"}/v1/:path*` },
      { source: "/api/notifications/:path*", destination: `${process.env.NEXT_PUBLIC_NOTIFICATIONS_URL || "http://notifications:4006"}/v1/:path*` },
      { source: "/api/sla/:path*", destination: `${process.env.NEXT_PUBLIC_SLA_URL || "http://sla:4011"}/v1/sla/:path*` },
      { source: "/api/onboarding/:path*", destination: `${process.env.NEXT_PUBLIC_ONBOARDING_URL || "http://onboarding:4009"}/v1/onboarding/:path*` },
      { source: "/api/autopilot/:path*", destination: `${process.env.NEXT_PUBLIC_AUTOPILOT_URL || "http://autopilot:4007"}/v1/:path*` },
    ];
  },
};
export default nextConfig;

import type { Metadata, Viewport } from "next";
import { Providers } from "@/components/layout/providers";
import { RegisterSW } from "@/components/pwa/register-sw";
import { Toaster } from "sonner";
import "./globals.css";

export const metadata: Metadata = {
  title: "ZU Seguridad Integral 360",
  description: "Plataforma Inteligente de Salud Ocupacional — Colombia",
  manifest: "/manifest.json",
  applicationName: "ZU 360",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
    title: "ZU 360",
  },
  icons: {
    icon: "/icons/icon-192.png",
    apple: "/icons/apple-touch-icon.png",
  },
};

export const viewport: Viewport = {
  themeColor: "#6BA31A",
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="es" suppressHydrationWarning>
      <body>
        <Providers>
          {children}
          <Toaster position="top-right" richColors closeButton />
        </Providers>
        <RegisterSW />
      </body>
    </html>
  );
}

import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ZU Portal Empresa Cliente",
  description: "Portal de gestión SST para empresas clientes",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (<html lang="es"><body>{children}</body></html>);
}

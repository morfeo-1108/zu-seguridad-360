import type { Config } from "tailwindcss";
const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        zu: {
          // Fondo claro corporativo (Notion/Figma) — conserva legibilidad
          bg: "#F6F7F4", surface: "#FFFFFF", surface2: "#F0F2ED",
          border: "#E3E7DD", text1: "#161A14", text2: "#535A4E", text3: "#969C8E",
          // Identidad de marca ZU 360
          accent: "#6BA31A",       // verde lima ZU (acento principal, versión legible en claro)
          lime: "#8BC34A",         // verde lima vivo del logo (highlights)
          graphite: "#12171F",     // grafito del fondo del logo (headers/sidebar oscuro)
          steel: "#6B7280",        // plateado metálico de la "U"
          green: "#16A34A", amber: "#D97706", red: "#DC2626",
        },
      },
      fontFamily: { sans: ["DM Sans", "sans-serif"], mono: ["DM Mono", "monospace"], display: ["var(--font-display)", "Georgia", "serif"] },
      boxShadow: {
        "zu-sm": "0 1px 2px rgba(0,0,0,0.06)",
        "zu": "0 1px 4px rgba(0,0,0,0.08), 0 0 0 1px rgba(0,0,0,0.04)",
        "zu-md": "0 4px 12px rgba(0,0,0,0.08), 0 0 0 1px rgba(0,0,0,0.04)",
      },
    },
  },
  plugins: [],
};
export default config;

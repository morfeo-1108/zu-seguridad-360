import { Fraunces } from "next/font/google";

// Serif display for the landing's headlines, against DM Sans in the body.
// Self-hosted by next/font (no request to Google at runtime) and scoped to the
// landing through a CSS variable, so the app screens do not download it.
export const display = Fraunces({
  subsets: ["latin"],
  axes: ["opsz"],
  display: "swap",
  variable: "--font-display",
});

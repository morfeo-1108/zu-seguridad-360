// Logo ZU 360 — reinterpretación del logo de marca como SVG vectorial.
// Concepto conservado: "ZU" con Z en verde lima y U en acero, "360" con arco,
// y el tagline "PREVENIMOS HOY, PROTEGEMOS SIEMPRE".
import React from "react";

export function ZULogo({ size = 40, dark = false, showText = false }: { size?: number; dark?: boolean; showText?: boolean }) {
  const lime = "#8BC34A";
  const limeDark = "#6BA31A";
  const steel = dark ? "#E2E5EA" : "#9CA3AF";
  return (
    <div style={{ display: "flex", alignItems: "center", gap: showText ? 10 : 0 }}>
      <svg width={size} height={size} viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="zuLime" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stopColor={lime} />
            <stop offset="100%" stopColor={limeDark} />
          </linearGradient>
          <linearGradient id="zuSteel" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stopColor={dark ? "#F1F3F6" : "#C4CAD2"} />
            <stop offset="100%" stopColor={steel} />
          </linearGradient>
        </defs>
        {/* Arco 360° */}
        <path d="M52 14 A26 26 0 1 1 20 10" stroke="url(#zuLime)" strokeWidth="3.5" strokeLinecap="round" fill="none" opacity="0.9" />
        <circle cx="53.5" cy="12.5" r="3.2" fill="url(#zuLime)" />
        {/* Z en verde lima */}
        <path d="M14 22 H30 L18 40 H31 V46 H12 L24 28 H14 Z" fill="url(#zuLime)" />
        {/* U en acero */}
        <path d="M33 22 H39 V38 A6 6 0 0 0 51 38 V22 H45 V38 A0 0 0 0 1 45 38 V22 H39" fill="none" />
        <path d="M34 20 V37 A7 7 0 0 0 48 37 V20 H42.5 V37 A1.5 1.5 0 0 1 39.5 37 V20 Z" fill="url(#zuSteel)" />
      </svg>
      {showText && (
        <div style={{ lineHeight: 1 }}>
          <div style={{ fontWeight: 800, fontSize: size * 0.34, letterSpacing: "-0.5px", color: dark ? "#fff" : "#161A14" }}>
            ZU <span style={{ color: limeDark }}>360</span>
          </div>
          <div style={{ fontSize: size * 0.14, fontWeight: 600, letterSpacing: "1px", color: dark ? "#8BC34A" : "#969C8E", marginTop: 2 }}>
            SEGURIDAD 360
          </div>
        </div>
      )}
    </div>
  );
}

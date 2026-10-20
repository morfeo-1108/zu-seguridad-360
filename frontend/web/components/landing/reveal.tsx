"use client";
import { ReactNode, useEffect, useRef } from "react";

// Fades a block in once it scrolls into view. One IntersectionObserver per
// instance is cheap at landing-page scale (a few dozen cards); unobserves
// itself after first reveal since content never needs to re-hide.
export function Reveal({ children, delay = 0, className }: { children: ReactNode; delay?: number; className?: string }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.classList.add("is-visible");
          observer.unobserve(el);
        }
      },
      { threshold: 0.15, rootMargin: "0px 0px -40px 0px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return (
    <div ref={ref} className={`reveal ${className || ""}`} style={{ transitionDelay: `${delay}ms` }}>
      {children}
    </div>
  );
}

// Tracks the cursor position within the element and exposes it as CSS
// custom properties (--mx/--my) that the .spotlight-card utility (see
// globals.css) uses to paint a glow following the pointer — the element
// itself does no painting, just the coordinate math.
export function spotlightHandler(e: React.MouseEvent<HTMLElement>) {
  const rect = e.currentTarget.getBoundingClientRect();
  e.currentTarget.style.setProperty("--mx", `${e.clientX - rect.left}px`);
  e.currentTarget.style.setProperty("--my", `${e.clientY - rect.top}px`);
}

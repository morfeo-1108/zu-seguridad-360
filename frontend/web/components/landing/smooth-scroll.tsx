"use client";
import { useEffect } from "react";
import { prefersReducedMotion } from "./motion";

// Lenis drives the scroll; GSAP's ticker drives Lenis, and every Lenis scroll
// updates ScrollTrigger — the documented integration, with one clock and no
// ScrollSmoother. Off entirely for reduced motion (native scroll).
export function SmoothScroll() {
  useEffect(() => {
    if (prefersReducedMotion()) return;
    let cleanup = () => {};
    let cancelled = false;
    (async () => {
      const [{ default: Lenis }, { gsap }, { ScrollTrigger }] = await Promise.all([
        import("lenis"), import("gsap"), import("gsap/ScrollTrigger"),
      ]);
      if (cancelled) return;
      gsap.registerPlugin(ScrollTrigger);
      const lenis = new Lenis({ lerp: 0.12, smoothWheel: true });
      lenis.on("scroll", ScrollTrigger.update);
      const tick = (time: number) => lenis.raf(time * 1000);
      gsap.ticker.add(tick);
      gsap.ticker.lagSmoothing(0);
      // In-page anchors (#capacidades) scroll through Lenis too.
      const onClick = (e: MouseEvent) => {
        const a = (e.target as HTMLElement).closest('a[href^="#"]') as HTMLAnchorElement | null;
        if (!a) return;
        const target = document.querySelector(a.getAttribute("href")!);
        if (target) { e.preventDefault(); lenis.scrollTo(target as HTMLElement, { offset: -72 }); }
      };
      document.addEventListener("click", onClick);
      cleanup = () => {
        document.removeEventListener("click", onClick);
        gsap.ticker.remove(tick);
        lenis.destroy();
      };
    })();
    return () => { cancelled = true; cleanup(); };
  }, []);
  return null;
}

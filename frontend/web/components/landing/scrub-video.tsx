"use client";
import { useEffect, useRef, useState } from "react";
import { prefersReducedMotion } from "./motion";

type Sources = { webm: string; mp4: string };
type Props = {
  /** Element that ScrollTrigger pins and measures (the hero section). */
  triggerRef: React.RefObject<HTMLElement | null>;
  poster: string;
  sources: { wide: Sources; narrow: Sources };
  className?: string;
};

// The video is scrubbed by scroll, never autoplayed: currentTime follows
// ScrollTrigger's progress (start 'top top', end '+=100%', scrub: true).
// Seeks are coalesced to one per animation frame and skipped while a seek is
// pending, so fast wheel spins do not queue dozens of decoder seeks.
//
// Loading: the poster is the LCP element; the video source is attached only
// after the page has loaded and the browser is idle. The pin exists from
// mount, so the video arriving never shifts the layout. With
// prefers-reduced-motion the poster stays and nothing is pinned or scrubbed.
export function ScrubVideo({ triggerRef, poster, sources, className }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const video = videoRef.current;
    const trigger = triggerRef.current;
    if (!video || !trigger || prefersReducedMotion()) return;
    let killed = false;
    let st: { kill: () => void; progress: number } | null = null;
    let raf = 0;

    const seek = () => {
      raf = 0;
      if (!st || video.readyState < 1 || !video.duration) return;
      if (video.seeking) { raf = requestAnimationFrame(seek); return; }
      const target = st.progress * video.duration;
      if (Math.abs(video.currentTime - target) > 1 / 60) video.currentTime = target;
    };
    const requestSeek = () => { if (!raf) raf = requestAnimationFrame(seek); };

    (async () => {
      const [{ gsap }, { ScrollTrigger }] = await Promise.all([import("gsap"), import("gsap/ScrollTrigger")]);
      if (killed) return;
      gsap.registerPlugin(ScrollTrigger);
      st = ScrollTrigger.create({
        trigger, start: "top top", end: "+=100%", scrub: true, pin: true, anticipatePin: 1,
        onUpdate: requestSeek,
      });
    })();

    const attach = () => {
      if (killed) return;
      const set = window.matchMedia("(max-width: 767px)").matches ? sources.narrow : sources.wide;
      video.src = video.canPlayType('video/webm; codecs="vp9"') ? set.webm : set.mp4;
      video.load();
    };
    const onData = () => { setReady(true); requestSeek(); };
    video.addEventListener("loadeddata", onData);
    const idle = (cb: () => void) => {
      const w = window as Window & { requestIdleCallback?: (cb: () => void, o?: { timeout: number }) => number };
      if (w.requestIdleCallback) w.requestIdleCallback(cb, { timeout: 2500 });
      else setTimeout(cb, 1200);
    };
    if (document.readyState === "complete") idle(attach);
    else window.addEventListener("load", () => idle(attach), { once: true });

    return () => {
      killed = true;
      cancelAnimationFrame(raf);
      st?.kill();
      video.removeEventListener("loadeddata", onData);
    };
  }, [triggerRef, sources]);

  return (
    <div className={className} aria-hidden>
      <img src={poster} alt="" fetchPriority="high" decoding="async" className="absolute inset-0 h-full w-full object-cover" />
      <video
        ref={videoRef}
        muted
        playsInline
        preload="none"
        disablePictureInPicture
        tabIndex={-1}
        className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-700 ${ready ? "opacity-100" : "opacity-0"}`}
      />
    </div>
  );
}

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
// ScrollTrigger's progress (start 'top top', scrub: true). On wide screens the
// hero is pinned for one viewport of scroll (end '+=100%'). On phones it is
// not pinned: the hero can be taller than the visible area (Safari's toolbars
// take ~180px), and a pinned section hides whatever is below the fold until
// the pin ends. There the video scrubs while the hero scrolls away
// (end 'bottom top').
// Seeks are coalesced to one per animation frame and skipped while a seek is
// pending, so fast wheel spins do not queue dozens of decoder seeks.
//
// Loading: the poster is the LCP element; the video source is attached only
// after the page has loaded and the browser is idle. The pin exists from
// mount, so the video arriving never shifts the layout. With
// prefers-reduced-motion the poster stays and nothing is pinned or scrubbed.
//
// Robustness, so the hero works wherever the page is opened:
// - if the chosen format fails to decode, the other one is tried (WebM <-> MP4);
// - if the browser refuses to load media on its own (iOS Low Power Mode, some
//   in-app browsers), the first tap or click loads and primes it: a user
//   gesture is always allowed to start media. Until then the poster shows.
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
      const pin = window.matchMedia("(min-width: 768px)").matches && trigger.offsetHeight <= window.innerHeight + 1;
      st = ScrollTrigger.create({
        trigger, start: "top top", end: pin ? "+=100%" : "bottom top", scrub: true, pin, anticipatePin: pin ? 1 : 0,
        onUpdate: requestSeek,
      });
    })();

    const apple = /Apple/.test(navigator.vendor);
    const set = window.matchMedia("(max-width: 767px)").matches ? sources.narrow : sources.wide;
    // Safari (macOS and iOS) gets H.264: hardware-decoded and seekable
    // frame by frame. Its VP9 support is newer and seeks poorly.
    const first = !apple && video.canPlayType('video/webm; codecs="vp9"') ? set.webm : set.mp4;
    const second = first === set.webm ? set.mp4 : set.webm;
    let attached = false;
    const attach = () => {
      if (killed || attached) return;
      attached = true;
      video.src = first;
      // Safari loads little more than metadata otherwise and the scrub shows
      // the poster; other browsers fetch ranges as the user scrolls, so they
      // keep preload="none" and do not spend 3-5 MB of mobile data up front.
      if (apple) video.preload = "auto";
      video.load();
    };
    // iOS may not paint seeked frames of a video that never played: start and
    // immediately pause it once (muted + playsInline is allowed; in Low Power
    // Mode it is refused and the poster stays). Only on Safari: in Chrome a
    // play() turns into a full download of the file.
    const prime = () => { video.play()?.then(() => video.pause()).catch(() => {}); };
    const onMeta = () => { if (apple) prime(); };
    const onData = () => { setReady(true); requestSeek(); removeGesture(); };
    const onError = () => {
      if (video.getAttribute("src") === second || !second) return; // both failed: the poster stays
      video.src = second;
      video.load();
    };
    // Fallback for browsers that will not load media without a gesture.
    const gestures = ["touchend", "click", "keydown"] as const;
    const onGesture = () => {
      if (video.readyState >= 2) return removeGesture();
      if (!attached) attach();
      else if (video.networkState !== HTMLMediaElement.NETWORK_LOADING) video.load();
      prime();
    };
    const removeGesture = () => gestures.forEach((g) => window.removeEventListener(g, onGesture));
    gestures.forEach((g) => window.addEventListener(g, onGesture, { passive: true }));
    video.addEventListener("loadedmetadata", onMeta);
    video.addEventListener("loadeddata", onData);
    video.addEventListener("error", onError);
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
      video.removeEventListener("loadedmetadata", onMeta);
      video.removeEventListener("error", onError);
      removeGesture();
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

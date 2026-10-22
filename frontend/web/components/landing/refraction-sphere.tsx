"use client";
import { useEffect, useRef, useState } from "react";
import type * as ThreeTypes from "three";
import { prefersReducedMotion } from "./motion";

function webglAvailable(): boolean {
  try {
    const c = document.createElement("canvas");
    return !!(c.getContext("webgl2") || c.getContext("webgl"));
  } catch {
    return false;
  }
}

// Paints the texture of one parallax plane: thin lime/graphite lines and
// rings on transparency, drawn on a 2D canvas (no image download).
function planeTexture(kind: "grid" | "rings" | "beams"): HTMLCanvasElement {
  // Opaque but dark strokes on full transparency: alphaTest discards the
  // empty texels, and the lines stay "opaque" so the transmission pass (which
  // only sees opaque objects) refracts them through the sphere.
  const c = document.createElement("canvas");
  c.width = c.height = 512;
  const g = c.getContext("2d")!;
  g.clearRect(0, 0, 512, 512);
  if (kind === "grid") {
    g.strokeStyle = "#1f3313";
    g.lineWidth = 2;
    for (let i = 0; i <= 512; i += 64) {
      g.beginPath(); g.moveTo(i, 0); g.lineTo(i, 512); g.stroke();
      g.beginPath(); g.moveTo(0, i); g.lineTo(512, i); g.stroke();
    }
  } else if (kind === "rings") {
    for (let r = 60; r < 256; r += 48) {
      g.strokeStyle = r % 96 === 12 ? "#4f7a1e" : "#2b4416";
      g.lineWidth = 4;
      g.beginPath(); g.arc(256, 256, r, 0, Math.PI * 2); g.stroke();
    }
  } else {
    g.fillStyle = "#8BC34A";
    for (let i = 0; i < 3; i++) g.fillRect(150 + i * 90, 0, 8, 512);
  }
  return c;
}

/**
 * Glass sphere with real refraction (decisión de negocio: siempre activa).
 * Spec: SphereGeometry(1, 64, 64); MeshPhysicalMaterial with transmission 1,
 * thickness 1.5, roughness 0.05, metalness 0, ior 1.45, iridescence 0.4,
 * iridescenceIOR 1.3; environment rendered from its own scene with a
 * CubeCamera; three background planes at z -8/-4/0 moving 0.2×/0.5×/1× the
 * scroll delta; sphere renderOrder 1; ACESFilmic tone mapping, exposure 1.1.
 *
 * The only fallback is technical: without WebGL a static gradient renders
 * instead. With prefers-reduced-motion one still frame is drawn (no spin,
 * no parallax). Frames are rendered on demand (pointer/scroll), never in a
 * continuous loop, and never off screen or in a hidden tab.
 */
export function RefractionSphere({ className }: { className?: string }) {
  const hostRef = useRef<HTMLDivElement>(null);
  const [noWebGL, setNoWebGL] = useState(false);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;
    if (!webglAvailable()) { setNoWebGL(true); return; }
    const reduced = prefersReducedMotion();
    let disposed = false;
    let cleanup = () => {};

    const whenIdle = (cb: () => void) => {
      const w = window as Window & { requestIdleCallback?: (cb: () => void, o?: { timeout: number }) => number };
      if (w.requestIdleCallback) w.requestIdleCallback(cb, { timeout: 3000 }); else setTimeout(cb, 1500);
    };
    // Everything (module download, environment capture, first frame) waits
    // for an idle moment after load, off the critical rendering path.
    const start = async () => {
      const THREE = await import("three");
      if (disposed) return;
      // No MSAA and pixel ratio 1: transmission already renders the scene twice
      // per frame; supersampling it costs more than it shows behind a glass ball.
      const renderer = new THREE.WebGLRenderer({ antialias: false, alpha: true, powerPreference: "high-performance" });
      renderer.setPixelRatio(1);
      renderer.toneMapping = THREE.ACESFilmicToneMapping;
      renderer.toneMappingExposure = 1.1;
      renderer.outputColorSpace = THREE.SRGBColorSpace;
      renderer.domElement.setAttribute("aria-hidden", "true");
      Object.assign(renderer.domElement.style, { width: "100%", height: "100%", display: "block" });
      host.appendChild(renderer.domElement);

      const scene = new THREE.Scene();
      const camera = new THREE.PerspectiveCamera(35, 1, 0.1, 100);
      camera.position.set(0, 0, 7);

      // Environment generated from a scene (not an HDR download): a dark
      // room with lime and white emissive panels, captured by a CubeCamera.
      const envScene = new THREE.Scene();
      envScene.background = new THREE.Color("#0b0f0c");
      const panel = (color: string, intensity: number, pos: [number, number, number], size: [number, number]) => {
        const m = new THREE.Mesh(new THREE.PlaneGeometry(...size), new THREE.MeshBasicMaterial({ color: new THREE.Color(color).multiplyScalar(intensity), side: THREE.DoubleSide }));
        m.position.set(...pos);
        m.lookAt(0, 0, 0);
        envScene.add(m);
      };
      panel("#8BC34A", 2.2, [6, 3, 4], [5, 3]);
      panel("#ffffff", 0.9, [-6, 4, 2], [3, 3]);
      panel("#16A34A", 1.4, [0, -6, -3], [8, 2]);
      panel("#ffffff", 0.6, [0, 2, -7], [10, 2]);
      const cubeRT = new THREE.WebGLCubeRenderTarget(64, { type: THREE.HalfFloatType, generateMipmaps: true, minFilter: THREE.LinearMipmapLinearFilter });
      const cubeCamera = new THREE.CubeCamera(0.1, 50, cubeRT);
      cubeCamera.update(renderer, envScene);
      scene.environment = cubeRT.texture;

      // Parallax planes behind the sphere (alphaTest keeps them opaque for
      // the transmission pass, so the sphere refracts their lines).
      const mkPlane = (kind: "grid" | "rings" | "beams", z: number, size: number, opacity: number) => {
        const tex = new THREE.CanvasTexture(planeTexture(kind));
        tex.colorSpace = THREE.SRGBColorSpace;
        const mat = new THREE.MeshBasicMaterial({ map: tex, alphaTest: 0.5, transparent: false, color: new THREE.Color(1, 1, 1).multiplyScalar(opacity) });
        const mesh = new THREE.Mesh(new THREE.PlaneGeometry(size, size), mat);
        mesh.position.z = z;
        scene.add(mesh);
        return mesh;
      };
      // Dark backdrop (hero color): the transmission pass needs something
      // opaque behind the sphere, or it samples an empty, bright target and
      // the glass reads as a milky ball. The radial mask fades its edges.
      const backdrop = new THREE.Mesh(new THREE.PlaneGeometry(60, 60), new THREE.MeshBasicMaterial({ color: "#0b0f0c" }));
      backdrop.position.z = -10;
      scene.add(backdrop);
      const planes = [
        { mesh: mkPlane("grid", -8, 26, 1), factor: 0.2 },
        { mesh: mkPlane("rings", -4, 8, 1), factor: 0.5 },
        { mesh: mkPlane("beams", 0, 3.2, 0.55), factor: 1 },
      ];
      planes[2].mesh.position.set(-0.2, 0.2, 0);

      const material = new THREE.MeshPhysicalMaterial({
        transmission: 1, thickness: 1.5, roughness: 0.05, metalness: 0, ior: 1.45,
        iridescence: 0.4, iridescenceIOR: 1.3, envMapIntensity: 1.2,
      });
      const sphere = new THREE.Mesh(new THREE.SphereGeometry(1, 64, 64), material);
      sphere.position.set(0, 0, 2);
      sphere.scale.setScalar(0.9);
      sphere.renderOrder = 1; // after what it refracts
      scene.add(sphere);

      const resize = () => {
        const w = host.clientWidth, h = host.clientHeight;
        if (!w || !h) return;
        renderer.setSize(w, h, false);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      resize();

      // Render on demand: one frame when idle, then animate (capped at 30 fps)
      // only while the pointer moves or the page scrolls, stopping shortly
      // after the last interaction. A continuous loop with transmission (the
      // scene is drawn twice per frame) is a long main-thread task on every
      // frame on devices without a fast GPU.
      let lastScroll = window.scrollY;
      let raf = 0;
      let visible = true;
      let activeUntil = 0;
      let lastFrame = 0;
      const pointer = { x: 0, y: 0 };
      const render = () => renderer.render(scene, camera);
      const loop = (t: number) => {
        raf = 0;
        if (!visible || document.hidden) return;
        if (t - lastFrame >= 33) {
          lastFrame = t;
          const y = window.scrollY;
          const delta = (y - lastScroll) * 0.004; // px → world units
          lastScroll = y;
          for (const p of planes) p.mesh.position.y += delta * p.factor;
          sphere.rotation.y += 0.004;
          sphere.rotation.x += (pointer.y * 0.25 - sphere.rotation.x) * 0.12;
          sphere.position.x += (pointer.x * 0.3 - sphere.position.x) * 0.12;
          render();
        }
        if (t < activeUntil) raf = requestAnimationFrame(loop);
      };
      const wake = () => {
        if (reduced) return;
        activeUntil = performance.now() + 1200;
        if (!raf && visible && !document.hidden) raf = requestAnimationFrame(loop);
      };
      const onPointer = (e: PointerEvent) => {
        pointer.x = (e.clientX / window.innerWidth) * 2 - 1;
        pointer.y = (e.clientY / window.innerHeight) * 2 - 1;
        wake();
      };
      // Visibility only gates rendering; it never starts an animation by itself.
      const io = new IntersectionObserver(([e]) => { visible = e.isIntersecting; lastScroll = window.scrollY; });
      const onVis = () => { if (!document.hidden) wake(); };
      let sized = false;
      const ro = new ResizeObserver(() => { if (!sized) { sized = true; return; } resize(); render(); });
      ro.observe(host);
      io.observe(host);

      // Compile shaders without blocking where the driver supports
      // KHR_parallel_shader_compile, then draw the first frame.
      await renderer.compileAsync(scene, camera);
      if (disposed) { io.disconnect(); ro.disconnect(); cubeRT.dispose(); renderer.dispose(); renderer.domElement.remove(); return; }
      render();
      if (!reduced) {
        window.addEventListener("pointermove", onPointer, { passive: true });
        window.addEventListener("scroll", wake, { passive: true });
        document.addEventListener("visibilitychange", onVis);
      }

      cleanup = () => {
        cancelAnimationFrame(raf);
        io.disconnect();
        ro.disconnect();
        window.removeEventListener("pointermove", onPointer);
        window.removeEventListener("scroll", wake);
        document.removeEventListener("visibilitychange", onVis);
        scene.traverse((o) => {
          const m = o as ThreeTypes.Mesh;
          m.geometry?.dispose();
          const mat = m.material as ThreeTypes.Material & { map?: ThreeTypes.Texture };
          mat?.map?.dispose();
          mat?.dispose?.();
        });
        envScene.traverse((o) => { (o as ThreeTypes.Mesh).geometry?.dispose(); ((o as ThreeTypes.Mesh).material as ThreeTypes.Material)?.dispose?.(); });
        cubeRT.dispose();
        renderer.dispose();
        renderer.domElement.remove();
      };
    };
    const kick = () => whenIdle(() => { void start(); });
    if (document.readyState === "complete") kick(); else window.addEventListener("load", kick, { once: true });

    return () => { disposed = true; window.removeEventListener("load", kick); cleanup(); };
  }, []);

  if (noWebGL) {
    return (
      <div className={className} aria-hidden>
        <div className="absolute left-1/2 top-1/2 aspect-square w-[min(60%,420px)] -translate-x-1/2 -translate-y-1/2 rounded-full bg-[radial-gradient(circle_at_35%_30%,rgba(255,255,255,0.55),rgba(139,195,74,0.35)_35%,rgba(22,163,74,0.15)_60%,transparent_72%)] shadow-[inset_0_0_60px_rgba(255,255,255,0.15)]" />
      </div>
    );
  }
  // Radial mask dissolves the canvas edges into the hero.
  return <div ref={hostRef} className={className} aria-hidden style={{ maskImage: "radial-gradient(closest-side, #000 62%, transparent 100%)", WebkitMaskImage: "radial-gradient(closest-side, #000 62%, transparent 100%)" }} />;
}

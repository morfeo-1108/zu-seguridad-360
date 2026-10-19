"use client";
import { useEffect, useState } from "react";
import { Download } from "lucide-react";

// Botón "Instalar" para escritorio/móvil. Aparece solo cuando el navegador
// determina que la PWA es instalable (evento beforeinstallprompt).
export function InstallButton() {
  const [prompt, setPrompt] = useState<any>(null);
  const [installed, setInstalled] = useState(false);

  useEffect(() => {
    const onPrompt = (e: any) => {
      e.preventDefault();
      setPrompt(e);
    };
    const onInstalled = () => { setInstalled(true); setPrompt(null); };
    window.addEventListener("beforeinstallprompt", onPrompt);
    window.addEventListener("appinstalled", onInstalled);
    return () => {
      window.removeEventListener("beforeinstallprompt", onPrompt);
      window.removeEventListener("appinstalled", onInstalled);
    };
  }, []);

  if (installed || !prompt) return null;

  const install = async () => {
    prompt.prompt();
    await prompt.userChoice;
    setPrompt(null);
  };

  return (
    <button onClick={install}
      className="flex items-center gap-1.5 text-[12px] font-semibold text-zu-accent bg-zu-surface2 border border-zu-border rounded-lg px-3 py-1.5 hover:bg-white transition"
      title="Instalar ZU 360 en tu escritorio">
      <Download size={13} /> Instalar app
    </button>
  );
}

"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/store";

export default function HomePage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  useEffect(() => { router.replace(isAuthenticated ? "/dashboard" : "/auth/login"); }, [isAuthenticated, router]);
  return <div className="min-h-screen flex items-center justify-center bg-zu-bg"><div className="text-zu-text2 text-sm">Cargando ZU 360...</div></div>;
}

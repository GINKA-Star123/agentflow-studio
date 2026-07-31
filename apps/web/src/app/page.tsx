"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuthStore } from "@/stores/auth-store";

export default function HomePage() {
  const router = useRouter();
  const status = useAuthStore((state) => state.status);
  const initialized = useAuthStore((state) => state.initialized);

  useEffect(() => {
    if (!initialized || status === "loading") {
      return;
    }

    if (status === "authenticated") {
      router.replace("/dashboard");
    } else {
      router.replace("/login");
    }
  }, [initialized, status, router]);

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 text-sm text-slate-500">
      正在进入系统...
    </main>
  );
}
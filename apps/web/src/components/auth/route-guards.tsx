"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuthStore } from "@/stores/auth-store";

type GuardProps = {
    children: React.ReactNode;
    redirectTo?: string;
};

function GuardLoading({message}:{message:string}){
    return(
        <main className = "flex min-h-screen items-center justify-center bg-slate-50 text-sm text-slate-500">
            {message}
        </main>
    )
}

export function AuthGuard({children,redirectTo = "/login"}:GuardProps) {
    const router = useRouter();
    const status = useAuthStore((state) => state.status);
    const initialized = useAuthStore((state) => state.initialized);

    useEffect(() => {
        if (initialized && status === "anonymous"){
            router.replace(redirectTo);
        }
    }, [initialized, status, router, redirectTo]);

    if (!initialized || status  ==="loading"){
        return <GuardLoading message = "正在检查登录态"></GuardLoading>;
    }

    if (status !== "authenticated"){
        return <GuardLoading message = "正在跳转到登录页"/>;
    }

    return <>{children}</>;
}

export function GuestGuard({children,redirectTo = "/dashboard"}:GuardProps) {
    const router = useRouter();
    const status = useAuthStore((state) => state.status);
    const initialized = useAuthStore((state) => state.initialized);

    useEffect(() => {
    if (initialized && status === "authenticated") {
      router.replace(redirectTo);
    }
  }, [initialized, status, redirectTo, router]);

  if (!initialized || status === "loading") {
    return <GuardLoading message="正在初始化..." />;
  }

  if (status === "authenticated") {
    return <GuardLoading message="正在进入工作台..." />;
  }

  return <>{children}</>;
}
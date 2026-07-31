"use client";

import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth-store";
import { WorkspaceSwitcher } from "@/components/app-shell/workspace-switcher";

export function AppTopbar() {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  function handleLogout() {
    logout();
    router.replace("/login");
  }

  return (
    <header className="sticky top-0 z-20 border-b border-slate-200 bg-white/95 backdrop-blur">
      <div className="flex h-16 items-center justify-between gap-4 px-4 sm:px-6">
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
            当前 Workspace
          </p>
          <div className="mt-1">
            <WorkspaceSwitcher />
          </div>
        </div>

        <div className="flex items-center gap-3">
          <div className="hidden text-right md:block">
            <p className="truncate text-sm font-medium text-slate-950">
              {user?.display_name ?? "用户"}
            </p>
            <p className="truncate text-xs text-slate-500">
              {user?.email ?? "-"}
            </p>
          </div>

          <Button variant="outline" className="gap-2" onClick={handleLogout}>
            <LogOut className="h-4 w-4" />
            <span>退出</span>
          </Button>
        </div>
      </div>
    </header>
  );
}
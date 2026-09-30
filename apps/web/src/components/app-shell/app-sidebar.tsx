"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Building2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { useAuthStore } from "@/stores/auth-store";
import { cn } from "@/lib/utils";
import { appNavItems } from "@/components/app-shell/navigation";

export function AppSidebar() {
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const currentWorkspace = useAuthStore((state) => state.currentWorkspace);

  return (
    <aside className="hidden h-screen w-72 shrink-0 border-r border-slate-200 bg-white lg:flex lg:flex-col">
      <div className="border-b border-slate-200 px-5 py-5">
        <p className="text-sm font-semibold text-slate-950">AgentFlow Studio</p>
        <p className="mt-1 text-xs text-slate-500">Workflow 工作台</p>
      </div>

      <div className="px-4 py-4">
        <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
          <div className="flex items-center gap-2 text-xs text-slate-500">
            <Building2 className="h-3.5 w-3.5" />
            <span>当前 Workspace</span>
          </div>
          <p className="mt-2 truncate text-sm font-medium text-slate-950">
            {currentWorkspace?.name ?? "未选择 Workspace"}
          </p>
          <div className="mt-2 flex items-center gap-2">
            <Badge variant="secondary" className="text-xs">
              {currentWorkspace?.role ?? "-"}
            </Badge>
          </div>
        </div>
      </div>

      <nav className="flex-1 px-3">
        <div className="mb-2 px-2 text-xs font-medium uppercase tracking-wide text-slate-400">
          基础导航
        </div>

        <div className="space-y-1">
          {appNavItems.map((item) => {
            const active =
              pathname === item.href || pathname.startsWith(`${item.href}/`);

            if (item.disabled) {
              return (
                <div
                  key={item.label}
                  className="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-slate-400"
                >
                  <item.icon className="h-4 w-4 shrink-0" />
                  <div className="min-w-0 flex-1">
                    <div className="truncate font-medium">{item.label}</div>
                    <div className="truncate text-xs text-slate-400">
                      {item.description}
                    </div>
                  </div>
                  <Badge variant="outline" className="text-[10px]">
                    待开放
                  </Badge>
                </div>
              );
            }

            return (
              <Link
                key={item.label}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm transition",
                  active
                    ? "bg-slate-950 text-white"
                    : "text-slate-700 hover:bg-slate-100",
                )}
              >
                <item.icon className="h-4 w-4 shrink-0" />
                <div className="min-w-0 flex-1">
                  <div className="truncate font-medium">{item.label}</div>
                  <div
                    className={cn(
                      "truncate text-xs",
                      active ? "text-slate-300" : "text-slate-500",
                    )}
                  >
                    {item.description}
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      </nav>

      <div className="border-t border-slate-200 px-5 py-4">
        <p className="text-xs text-slate-400">当前用户</p>
        <p className="mt-1 truncate text-sm font-medium text-slate-950">
          {user?.display_name ?? user?.email ?? "-"}
        </p>
        <p className="mt-1 truncate text-xs text-slate-500">
          {user?.email ?? "-"}
        </p>
      </div>
    </aside>
  );
}
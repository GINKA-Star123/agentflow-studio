"use client";

import { Check, ChevronDown, Building2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuthStore } from "@/stores/auth-store";
import { Badge } from "@/components/ui/badge";

export function WorkspaceSwitcher() {
  const workspaces = useAuthStore((state) => state.workspaces);
  const currentWorkspace = useAuthStore((state) => state.currentWorkspace);
  const setCurrentWorkspace = useAuthStore((state) => state.setCurrentWorkspace);

  if (workspaces.length === 0) {
    return (
      <Button variant="outline" className="h-10 gap-2" disabled>
        <Building2 className="h-4 w-4" />
        <span>暂无 Workspace</span>
      </Button>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger>
        <Button variant="outline" className="h-10 min-w-[220px] justify-between gap-3">
          <span className="flex min-w-0 items-center gap-2">
            <Building2 className="h-4 w-4 shrink-0" />
            <span className="truncate">
              {currentWorkspace?.name ?? workspaces[0]?.name ?? "未选择 Workspace"}
            </span>
          </span>
          <ChevronDown className="h-4 w-4 shrink-0 opacity-60" />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="start" className="w-80">
        <DropdownMenuLabel>Workspace 切换</DropdownMenuLabel>
        <DropdownMenuSeparator />

        {workspaces.map((workspace) => {
          const active = workspace.id === currentWorkspace?.id;

          return (
            <DropdownMenuItem
              key={workspace.id}
              onSelect={() => setCurrentWorkspace(workspace.id)}
              className="cursor-pointer"
            >
              <div className="flex w-full items-center justify-between gap-3">
                <div className="min-w-0">
                  <div className="truncate text-sm font-medium">{workspace.name}</div>
                  <div className="text-xs text-slate-500">{workspace.role}</div>
                </div>

                {active ? <Check className="h-4 w-4 shrink-0" /> : null}
              </div>
            </DropdownMenuItem>
          );
        })}

        <DropdownMenuSeparator />
        <div className="px-2 pb-1 text-xs text-slate-500">
          当前角色：{currentWorkspace?.role ?? "-"}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
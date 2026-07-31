"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useAuthStore } from "@/stores/auth-store";

export default function DashboardPage() {
  const user = useAuthStore((state) => state.user);
  const workspaces = useAuthStore((state) => state.workspaces);
  const currentWorkspace = useAuthStore((state) => state.currentWorkspace);

  return (
    <div className="grid gap-4 xl:grid-cols-3">
      <Card className="xl:col-span-2">
        <CardHeader>
          <CardTitle>工作台概览</CardTitle>
          <CardDescription>当前账号与 Workspace 状态。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <div>
            <p className="text-slate-500">用户邮箱</p>
            <p className="font-medium text-slate-950">{user?.email ?? "-"}</p>
          </div>
          <div>
            <p className="text-slate-500">显示名称</p>
            <p className="font-medium text-slate-950">{user?.display_name ?? "-"}</p>
          </div>
          <div>
            <p className="text-slate-500">当前 Workspace</p>
            <p className="font-medium text-slate-950">
              {currentWorkspace?.name ?? "-"}
            </p>
          </div>
          <div>
            <p className="text-slate-500">当前角色</p>
            <Badge variant="secondary" className="mt-1">
              {currentWorkspace?.role ?? "-"}
            </Badge>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Workspace 数量</CardTitle>
          <CardDescription>你当前可访问的工作区。</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-3xl font-semibold text-slate-950">
            {workspaces.length}
          </div>
        </CardContent>
      </Card>

      <Card className="xl:col-span-3">
        <CardHeader>
          <CardTitle>下一步</CardTitle>
          <CardDescription>Phase 3 后续批次将进入 Workflow Designer。</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-slate-600">
          目前这里是工作台入口。下一批会开始铺 Workflow Designer 页面骨架、节点面板和画布区域。
        </CardContent>
      </Card>
    </div>
  );
}
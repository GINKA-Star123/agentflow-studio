import type { LucideIcon } from "lucide-react";
import { BookOpen, LayoutDashboard, Settings2, Users, Workflow } from "lucide-react";

export type AppNavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
  description: string;
  disabled?: boolean;
};

export const appNavItems: AppNavItem[] = [
  {
    label: "工作台",
    href: "/dashboard",
    icon: LayoutDashboard,
    description: "总览与快捷入口",
  },
  {
    label: "工作流",
    href: "/dashboard/workflows",
    icon: Workflow,
    description: "Workflow Designer",
  },
  {
    label: "知识库",
    href: "/dashboard/knowledge-bases",
    icon: BookOpen,
    description: "文档与检索",
    disabled: true,
  },
  {
    label: "成员",
    href: "/dashboard/members",
    icon: Users,
    description: "Workspace 成员",
    disabled: true,
  },
  {
    label: "设置",
    href: "/dashboard/settings",
    icon: Settings2,
    description: "系统配置",
    disabled: true,
  },
];
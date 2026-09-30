import { AuthGuard } from "@/components/auth/route-guards";
import { AppShell } from "@/components/app-shell/app-shell";

type DashboardLayoutProps = {
  children: React.ReactNode;
};

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  return (
    <AuthGuard>
      <AppShell>{children}</AppShell>
    </AuthGuard>
  );
}
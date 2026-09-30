"use client";

import { AuthForm } from "@/components/auth/auth-form";
import { GuestGuard } from "@/components/auth/route-guards";

export default function RegisterPage() {
  return (
    <GuestGuard>
      <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4 py-10">
        <AuthForm
          mode="register"
          title="注册"
          description="创建账号后会自动生成默认 Workspace。"
          submitText="注册"
          switchHref="/login"
          switchText="已有账号？去登录"
        />
      </main>
    </GuestGuard>
  );
}
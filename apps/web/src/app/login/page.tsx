"use client";

import { AuthForm } from "@/components/auth/auth-form";
import { GuestGuard } from "@/components/auth/route-guards";

export default function LoginPage(){
    return (
        <GuestGuard>
            <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4 py-10">
                <AuthForm
                    mode = "login"
                    title = "登录"
                    description = "登录以使用 AgentFlow Studio"
                    submitText = "登录"
                    switchHref = "/register"
                    switchText = "注册"
                 />
            </main>
        </GuestGuard>
    )
}
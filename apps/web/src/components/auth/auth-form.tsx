"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth-store";

type AuthFormProps = {
    mode : "login" | "register";
    title : string;
    description : string;
    submitText:string;
    switchHref : string;
    switchText : string;
};

export function AuthForm({
    mode,
    title,
    description,
    submitText,
    switchHref,
    switchText,
}: AuthFormProps) {
    const router = useRouter();

    const status = useAuthStore((state) =>state.status);
    const errorMessage = useAuthStore((state)=>state.errorMessage);
    const login = useAuthStore((state)=>state.login);
    const register = useAuthStore((state)=>state.register);

    const [email,setEmail] = useState("");
    const [password,setPassword] = useState("");
    const [displayName, setDisplayName] = useState("");
    const [workspaceName, setWorkspaceName] = useState("");

    const isSubmitting = status === "loading";

    async function handleSubmit(event:FormEvent<HTMLFormElement>){
        event.preventDefault();

        try {
            if (mode === "login"){
                await login({
                    email:email.trim(),
                    password,
                });
            }else{
                await register({
                    email:email.trim(),
                    password,
                    display_name:displayName.trim()||undefined,
                    workspace_name:workspaceName.trim()||undefined,
                });
            }

            router.replace("/dashboard");
        } catch{
            return ;
        }
    }
    return (
    <Card className="w-full max-w-md shadow-sm">
      <CardHeader>
        <CardTitle className="text-2xl">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>

      <CardContent>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="email">邮箱</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="user@example.com"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
          </div>

          {mode === "register" ? (
            <>
              <div className="space-y-2">
                <Label htmlFor="display-name">显示名称</Label>
                <Input
                  id="display-name"
                  type="text"
                  autoComplete="name"
                  placeholder="你的名称"
                  value={displayName}
                  onChange={(event) => setDisplayName(event.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="workspace-name">Workspace 名称</Label>
                <Input
                  id="workspace-name"
                  type="text"
                  placeholder="我的 Workspace"
                  value={workspaceName}
                  onChange={(event) => setWorkspaceName(event.target.value)}
                />
              </div>
            </>
          ) : null}

          <div className="space-y-2">
            <Label htmlFor="password">密码</Label>
            <Input
              id="password"
              type="password"
              autoComplete={mode === "login" ? "current-password" : "new-password"}
              placeholder="请输入密码"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </div>

          {errorMessage ? (
            <div className="rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
              {errorMessage}
            </div>
          ) : null}

          <Button className="w-full" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "处理中..." : submitText}
          </Button>
        </form>

        <div className="mt-4 text-center text-sm text-slate-600">
          <Link href={switchHref} className="underline-offset-4 hover:underline">
            {switchText}
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}

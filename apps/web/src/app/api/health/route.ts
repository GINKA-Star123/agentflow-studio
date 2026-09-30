import {NextResponse} from "next/server";

import {
    buildWebHealth,
    checkHttpService,
    getOverallState,
    type HealthOverview,
} from "@/lib/health";

export const dynamic = "force-dynamic";

export async function GET(){
    const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";
    const aiRuntimeBaseUrl = process.env.AI_RUNTIME_BASE_URL ?? "http://localhost:8090";

    const services = await Promise.all([
        Promise.resolve(buildWebHealth()),
        checkHttpService({
            key:"api",
            label:"GO API 服务",
            baseUrl: apiBaseUrl,
            path: "/readyz",
        }),
        checkHttpService({
            key:"aiRuntime",
            label:"AI Runtime 服务",
            baseUrl: aiRuntimeBaseUrl,
            path: "/readyz",
        }),
    ]);

    const overview: HealthOverview = {
        overall: getOverallState(services),
        generatedAt: new Date().toISOString(),
        services,
    };

    return NextResponse.json(overview);
}
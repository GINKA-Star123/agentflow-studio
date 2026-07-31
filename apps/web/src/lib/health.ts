export type ServiceKey = "web" | "api" | "aiRuntime";

export type HealthState = "healthy" | "degraded" | "down";

export type ServiceHealth = {
  key: ServiceKey;
  label: string;
  target: string;
  state: HealthState;
  message: string;
  checkedAt: string;
  responseTimeMs?: number;
  statusCode?: number;
  detail?: unknown;
};

export type HealthOverview = {
  overall: HealthState;
  generatedAt: string;
  services: ServiceHealth[];
};

type CheckServiceInput = {
  key: ServiceKey;
  label: string;
  baseUrl: string;
  path: string;
};

export function buildWebHealth(): ServiceHealth {
  return {
    key: "web",
    label: "Web 前端",
    target: "Next.js",
    state: "healthy",
    message: "前端服务运行正常",
    checkedAt: new Date().toISOString(),
    responseTimeMs: 0,
    statusCode: 200,
  };
}

export async function checkHttpService(input: CheckServiceInput): Promise<ServiceHealth> {
  const startedAt = Date.now();
  const checkedAt = new Date().toISOString();
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 3000);

  const target = `${input.baseUrl}${input.path}`;

  try {
    const response = await fetch(target, {
      method: "GET",
      cache: "no-store",
      signal: controller.signal,
    });

    const responseTimeMs = Date.now() - startedAt;
    const detail = await readJsonSafely(response);

    return {
      key: input.key,
      label: input.label,
      target,
      state: response.ok ? "healthy" : "down",
      message: response.ok ? "服务运行正常" : `服务返回异常状态码：${response.status}`,
      checkedAt,
      responseTimeMs,
      statusCode: response.status,
      detail,
    };
  } catch (error) {
    const responseTimeMs = Date.now() - startedAt;

    return {
      key: input.key,
      label: input.label,
      target,
      state: "down",
      message: error instanceof Error ? error.message : "服务请求失败",
      checkedAt,
      responseTimeMs,
    };
  } finally {
    clearTimeout(timeout);
  }
}

export function getOverallState(services: ServiceHealth[]): HealthState {
  const downCount = services.filter((service) => service.state === "down").length;

  if (downCount === 0) {
    return "healthy";
  }

  if (downCount === services.length) {
    return "down";
  }

  return "degraded";
}

async function readJsonSafely(response: Response): Promise<unknown> {
  const contentType = response.headers.get("content-type") ?? "";

  if (!contentType.includes("application/json")) {
    return null;
  }

  try {
    return await response.json();
  } catch {
    return null;
  }
}
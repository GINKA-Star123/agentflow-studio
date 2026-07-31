import type {
  ApiResponse,
  AuthResult,
  LoginPayload,
  MeResult,
  RegisterPayload,
} from "@/types/auth";
import { clearAccessToken, getAccessToken } from "@/lib/auth-token";

const API_BASE_PATH = process.env.NEXT_PUBLIC_API_BASE_PATH ?? "/api/v1";

type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  auth?: boolean;
};

export class ApiClientError extends Error {
  status: number;
  code: string;
  details?: unknown;
  requestId?: string;

  constructor(input: {
    status: number;
    code: string;
    message: string;
    details?: unknown;
    requestId?: string;
  }) {
    super(input.message);
    this.name = "ApiClientError";
    this.status = input.status;
    this.code = input.code;
    this.details = input.details;
    this.requestId = input.requestId;
  }
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const method = options.method ?? "GET";
  const headers = new Headers();

  headers.set("Accept", "application/json");

  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  if (options.auth !== false) {
    const token = getAccessToken();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  const response = await fetch(`${API_BASE_PATH}${path}`, {
    method,
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
    cache: "no-store",
  });

  const payload = (await readJsonSafely(response)) as ApiResponse<T> | null;

  if (!response.ok) {
    if (response.status === 401) {
      clearAccessToken();
    }

    throw new ApiClientError({
      status: response.status,
      code: payload?.error?.code ?? "HTTP_ERROR",
      message: payload?.error?.message ?? `请求失败：${response.status}`,
      details: payload?.error?.details,
      requestId: payload?.request_id,
    });
  }

  if (!payload || payload.data === undefined) {
    throw new ApiClientError({
      status: response.status,
      code: "EMPTY_RESPONSE",
      message: "接口返回数据为空",
      requestId: payload?.request_id,
    });
  }

  return payload.data;
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

export const authApi = {
  register(payload: RegisterPayload) {
    return apiRequest<AuthResult>("/auth/register", {
      method: "POST",
      body: payload,
      auth: false,
    });
  },

  login(payload: LoginPayload) {
    return apiRequest<AuthResult>("/auth/login", {
      method: "POST",
      body: payload,
      auth: false,
    });
  },

  me() {
    return apiRequest<MeResult>("/auth/me", {
      method: "GET",
      auth: true,
    });
  },
};
import { create } from "zustand";

import { authApi } from "@/lib/api-client";
import { clearAccessToken, getAccessToken, setAccessToken } from "@/lib/auth-token";
import type { AuthResult, LoginPayload, MeResult, RegisterPayload, User } from "@/types/auth";
import type { WorkspaceSummary } from "@/types/workspace";

type AuthStatus = "idle" | "loading" | "authenticated" | "anonymous";

type AuthState = {
  status: AuthStatus;
  user: User | null;
  workspaces: WorkspaceSummary[];
  currentWorkspace: WorkspaceSummary | null;
  errorMessage: string;
  initialized: boolean;

  initialize: () => Promise<void>;
  register: (payload: RegisterPayload) => Promise<void>;
  login: (payload: LoginPayload) => Promise<void>;
  refreshMe: () => Promise<void>;
  logout: () => void;
  setCurrentWorkspace: (workspaceId: string) => void;
};

export const useAuthStore = create<AuthState>((set, get) => ({
  status: "idle",
  user: null,
  workspaces: [],
  currentWorkspace: null,
  errorMessage: "",
  initialized: false,

  async initialize() {
    if (get().initialized) {
      return;
    }

    const token = getAccessToken();
    if (!token) {
      setAnonymous(set);
      return;
    }

    set({ status: "loading", errorMessage: "" });

    try {
      const result = await authApi.me();
      setMeResult(set, result);
    } catch (error) {
      clearAccessToken();
      set({
        status: "anonymous",
        user: null,
        workspaces: [],
        currentWorkspace: null,
        errorMessage: getErrorMessage(error),
        initialized: true,
      });
    }
  },

  async register(payload) {
    set({ status: "loading", errorMessage: "" });

    try {
      const result = await authApi.register(payload);
      setAuthResult(set, result);
    } catch (error) {
      set({
        status: "anonymous",
        errorMessage: getErrorMessage(error),
        initialized: true,
      });
      throw error;
    }
  },

  async login(payload) {
    set({ status: "loading", errorMessage: "" });

    try {
      const result = await authApi.login(payload);
      setAuthResult(set, result);
    } catch (error) {
      set({
        status: "anonymous",
        errorMessage: getErrorMessage(error),
        initialized: true,
      });
      throw error;
    }
  },

  async refreshMe() {
    set({ status: "loading", errorMessage: "" });

    try {
      const result = await authApi.me();
      setMeResult(set, result);
    } catch (error) {
      clearAccessToken();
      setAnonymous(set, getErrorMessage(error));
      throw error;
    }
  },

  logout() {
    clearAccessToken();
    setAnonymous(set);
  },

  setCurrentWorkspace(workspaceId) {
    const workspace = get().workspaces.find((item) => item.id === workspaceId) ?? null;
    set({ currentWorkspace: workspace });
  },
}));

function setAuthResult(
  set: (state: Partial<AuthState>) => void,
  result: AuthResult,
) {
  setAccessToken(result.access_token);

  set({
    status: "authenticated",
    user: result.user,
    workspaces: result.workspaces,
    currentWorkspace: result.current_workspace ?? result.workspaces[0] ?? null,
    errorMessage: "",
    initialized: true,
  });
}

function setMeResult(
  set: (state: Partial<AuthState>) => void,
  result: MeResult,
) {
  set({
    status: "authenticated",
    user: result.user,
    workspaces: result.workspaces,
    currentWorkspace: result.current_workspace ?? result.workspaces[0] ?? null,
    errorMessage: "",
    initialized: true,
  });
}

function setAnonymous(
  set: (state: Partial<AuthState>) => void,
  errorMessage = "",
) {
  set({
    status: "anonymous",
    user: null,
    workspaces: [],
    currentWorkspace: null,
    errorMessage,
    initialized: true,
  });
}

function getErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }

  return "请求失败";
}
import type { WorkspaceSummary } from "@/types/workspace";

export type User = {
    id:string;
    email:string;
    display_name:string;
    status:"active"|"disabled";
};

export type AuthResult = {
    user:User;
    access_token:string;
    token_type:"Bearer";
    expires_at:string;
    current_workspace?:WorkspaceSummary|null;
    workspaces:WorkspaceSummary[];
};

export type MeResult = {
    user:User;
    current_workspace?:WorkspaceSummary|null;
    workspaces:WorkspaceSummary[];
};

export type RegisterPayload = {
    email:string;
    password:string;
    display_name?:string;
    workspace_name?:string;
}

export type LoginPayload = {
    email:string;
    password:string;
};

export type ApiErrorBody = {
    code:string;
    message:string;
    details?:unknown;
}

export type ApiResponse<T> = {
    data?:T;
    error?:ApiErrorBody;
    request_id:string;
};
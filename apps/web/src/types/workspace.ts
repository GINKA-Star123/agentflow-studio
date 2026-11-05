export type WorkspaceRole = "owner" | "admin" | "member" | "viewer";

export type WorkspaceSummary = {
    id:string;
    name:string ;
    owner_id :string;
    role:WorkspaceRole;
};

export type WorkspaceMember = {
    user_id:string;
    email:string;
    display_name:string;
    role:WorkspaceRole;
    joined_at:string;
}

export type WorkspaceListResponse = {
    items:WorkspaceSummary[];
}

export type WorkspaceMemberListResponse = {
    items:WorkspaceMember[];
}
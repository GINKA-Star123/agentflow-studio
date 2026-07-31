DROP INDEX IF EXISTS idx_workspace_members_role;
DROP INDEX IF EXISTS idx_workspace_members_user_id;
DROP INDEX IF EXISTS idx_workspace_members_workspace_id;
DROP TABLE IF EXISTS workspace_members;

DROP INDEX IF EXISTS idx_workspaces_deleted_at;
DROP INDEX IF EXISTS idx_workspaces_owner_id;
DROP TABLE IF EXISTS workspaces;

DROP INDEX IF EXISTS idx_users_email_lower;
DROP TABLE IF EXISTS users;
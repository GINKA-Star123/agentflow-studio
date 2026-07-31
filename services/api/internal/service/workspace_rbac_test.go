package service

import (
	"testing"

	"agentflow-studio/services/api/internal/model"

	"github.com/google/uuid"
)

func TestCanUpdateWorkspaceMemberRole(t *testing.T) {
	t.Run("owner 可以修改 admin 为 viewer", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleOwner)
		target := testMember(model.WorkspaceRoleAdmin)

		err := canUpdateWorkspaceMemberRole(actor, target, model.WorkspaceRoleViewer)
		if err != nil {
			t.Fatalf("owner 修改 admin 为 viewer 应该允许: %v", err)
		}
	})

	t.Run("owner 不能设置其他人为 owner", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleOwner)
		target := testMember(model.WorkspaceRoleMember)

		err := canUpdateWorkspaceMemberRole(actor, target, model.WorkspaceRoleOwner)
		if err == nil {
			t.Fatal("owner 不应该通过该接口设置其他人为 owner")
		}
	})

	t.Run("admin 可以修改 member 为 viewer", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleAdmin)
		target := testMember(model.WorkspaceRoleMember)

		err := canUpdateWorkspaceMemberRole(actor, target, model.WorkspaceRoleViewer)
		if err != nil {
			t.Fatalf("admin 修改 member 为 viewer 应该允许: %v", err)
		}
	})

	t.Run("admin 不能修改其他 admin", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleAdmin)
		target := testMember(model.WorkspaceRoleAdmin)

		err := canUpdateWorkspaceMemberRole(actor, target, model.WorkspaceRoleMember)
		if err == nil {
			t.Fatal("admin 不应该修改其他 admin")
		}
	})

	t.Run("member 不能修改成员角色", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleMember)
		target := testMember(model.WorkspaceRoleViewer)

		err := canUpdateWorkspaceMemberRole(actor, target, model.WorkspaceRoleMember)
		if err == nil {
			t.Fatal("member 不应该修改成员角色")
		}
	})

	t.Run("不能修改自己的角色", func(t *testing.T) {
		userID := uuid.New()

		actor := &model.WorkspaceMember{
			UserID: userID,
			Role:   model.WorkspaceRoleOwner,
		}

		target := &model.WorkspaceMember{
			UserID: userID,
			Role:   model.WorkspaceRoleOwner,
		}

		err := canUpdateWorkspaceMemberRole(actor, target, model.WorkspaceRoleViewer)
		if err == nil {
			t.Fatal("不应该允许修改自己的角色")
		}
	})
}

func TestCanRemoveWorkspaceMember(t *testing.T) {
	t.Run("owner 可以移除 admin", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleOwner)
		target := testMember(model.WorkspaceRoleAdmin)

		err := canRemoveWorkspaceMember(actor, target)
		if err != nil {
			t.Fatalf("owner 移除 admin 应该允许: %v", err)
		}
	})

	t.Run("owner 不能移除 owner", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleOwner)
		target := testMember(model.WorkspaceRoleOwner)

		err := canRemoveWorkspaceMember(actor, target)
		if err == nil {
			t.Fatal("不应该允许移除 owner")
		}
	})

	t.Run("admin 可以移除 member", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleAdmin)
		target := testMember(model.WorkspaceRoleMember)

		err := canRemoveWorkspaceMember(actor, target)
		if err != nil {
			t.Fatalf("admin 移除 member 应该允许: %v", err)
		}
	})

	t.Run("admin 不能移除其他 admin", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleAdmin)
		target := testMember(model.WorkspaceRoleAdmin)

		err := canRemoveWorkspaceMember(actor, target)
		if err == nil {
			t.Fatal("admin 不应该移除其他 admin")
		}
	})

	t.Run("viewer 不能移除成员", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleViewer)
		target := testMember(model.WorkspaceRoleMember)

		err := canRemoveWorkspaceMember(actor, target)
		if err == nil {
			t.Fatal("viewer 不应该移除成员")
		}
	})

	t.Run("不能移除自己", func(t *testing.T) {
		userID := uuid.New()

		actor := &model.WorkspaceMember{
			UserID: userID,
			Role:   model.WorkspaceRoleOwner,
		}

		target := &model.WorkspaceMember{
			UserID: userID,
			Role:   model.WorkspaceRoleOwner,
		}

		err := canRemoveWorkspaceMember(actor, target)
		if err == nil {
			t.Fatal("不应该允许移除自己")
		}
	})
}

func TestCanAddWorkspaceMember(t *testing.T) {
	t.Run("owner 可以添加 member", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleOwner)

		err := canAddWorkspaceMember(actor, model.WorkspaceRoleMember)
		if err != nil {
			t.Fatalf("owner 添加 member 应该允许: %v", err)
		}
	})

	t.Run("admin 可以添加 viewer", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleAdmin)

		err := canAddWorkspaceMember(actor, model.WorkspaceRoleViewer)
		if err != nil {
			t.Fatalf("admin 添加 viewer 应该允许: %v", err)
		}
	})

	t.Run("member 不能添加成员", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleMember)

		err := canAddWorkspaceMember(actor, model.WorkspaceRoleViewer)
		if err == nil {
			t.Fatal("member 不应该添加成员")
		}
	})

	t.Run("viewer 不能添加成员", func(t *testing.T) {
		actor := testMember(model.WorkspaceRoleViewer)

		err := canAddWorkspaceMember(actor, model.WorkspaceRoleMember)
		if err == nil {
			t.Fatal("viewer 不应该添加成员")
		}
	})
}

func testMember(role model.WorkspaceRole) *model.WorkspaceMember {
	return &model.WorkspaceMember{
		UserID: uuid.New(),
		Role:   role,
	}
}

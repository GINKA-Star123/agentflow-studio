package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTManagerGenerateAndVerifyAccessToken(t *testing.T) {
	manager, err := NewJWTManager(JWTConfig{
		Secret:         "test-secret",
		Issuer:         "agentflow-studio-test",
		AccessTokenTTL: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewJWTManager 返回错误: %v", err)
	}

	userID := uuid.New()
	email := "jwt-user@example.com"

	token, expiresAt, err := manager.GenerateAccessToken(userID, email)
	if err != nil {
		t.Fatalf("GenerateAccessToken 返回错误: %v", err)
	}

	if token == "" {
		t.Fatal("token 不能为空")
	}

	if time.Now().After(expiresAt) {
		t.Fatal("expiresAt 不应该早于当前时间")
	}

	claims, err := manager.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken 返回错误: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("claims.UserID 不一致，got=%s want=%s", claims.UserID, userID)
	}

	if claims.Email != email {
		t.Fatalf("claims.Email 不一致，got=%s want=%s", claims.Email, email)
	}

	if claims.Type != TokenTypeAccess {
		t.Fatalf("claims.Type 不一致，got=%s want=%s", claims.Type, TokenTypeAccess)
	}
}

func TestJWTManagerRejectsInvalidToken(t *testing.T) {
	manager, err := NewJWTManager(JWTConfig{
		Secret:         "test-secret",
		Issuer:         "agentflow-studio-test",
		AccessTokenTTL: 2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewJWTManager 返回错误: %v", err)
	}

	_, err = manager.VerifyAccessToken("invalid-token")
	if err == nil {
		t.Fatal("无效 token 不应该校验通过")
	}
}

func TestExtractBearerToken(t *testing.T) {
	token, err := ExtractBearerToken("Bearer abc.def.ghi")
	if err != nil {
		t.Fatalf("ExtractBearerToken 返回错误: %v", err)
	}

	if token != "abc.def.ghi" {
		t.Fatalf("token 不一致，got=%s want=%s", token, "abc.def.ghi")
	}
}

func TestExtractBearerTokenRejectsInvalidHeader(t *testing.T) {
	cases := []string{
		"",
		"abc.def.ghi",
		"Basic abc.def.ghi",
		"Bearer",
		"Bearer ",
	}

	for _, item := range cases {
		_, err := ExtractBearerToken(item)
		if err == nil {
			t.Fatalf("非法 Authorization 应该被拒绝: %q", item)
		}
	}
}

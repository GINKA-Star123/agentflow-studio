package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordAndComparePassword(t *testing.T) {
	password := "password123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword 返回错误: %v", err)
	}

	if hash == "" {
		t.Fatal("密码哈希不能为空")
	}

	if hash == password {
		t.Fatal("密码哈希不能等于明文密码")
	}

	if err := ComparePassword(hash, password); err != nil {
		t.Fatalf("正确密码应该校验通过: %v", err)
	}
}

func TestComparePasswordWithWrongPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword 返回错误: %v", err)
	}

	err = ComparePassword(hash, "wrong-password")
	if err == nil {
		t.Fatal("错误密码不应该校验通过")
	}
}

func TestValidatePasswordRejectsWeakPassword(t *testing.T) {
	err := ValidatePassword("1234567")
	if err == nil {
		t.Fatal("少于 8 位的密码应该被拒绝")
	}
}

func TestValidatePasswordRejectsTooLongPassword(t *testing.T) {
	password := strings.Repeat("a", MaxBcryptPasswordBytes+1)

	err := ValidatePassword(password)
	if err == nil {
		t.Fatal("超过 bcrypt 72 字节限制的密码应该被拒绝")
	}
}

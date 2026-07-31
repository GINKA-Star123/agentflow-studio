package model

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type User struct {
	BaseModel

	Email        string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_email" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName  string     `gorm:"type:varchar(100);not null" json:"display_name"`
	Status       UserStatus `gorm:"type:varchar(32);not null;default:'active'" json:"status"`
}

func (User) TableName() string {
	return "users"
}

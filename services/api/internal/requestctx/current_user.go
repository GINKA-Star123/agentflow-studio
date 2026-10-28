package requestctx

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CurrentUserContextKey = "current_user"

var ErrCurrentUserNotFound = errors.New("current user not found")

type CurrentUser struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func SetCurrentUser(c *gin.Context, user CurrentUser) {
	c.Set(CurrentUserContextKey, user)
}

func GetCurrentUser(c *gin.Context) (*CurrentUser, error) {
	value, exists := c.Get(CurrentUserContextKey)
	if !exists {
		return nil, ErrCurrentUserNotFound
	}

	user, ok := value.(CurrentUser)
	if !ok {
		return nil, ErrCurrentUserNotFound
	}

	return &user, nil
}

func MustGetCurrentUser(c *gin.Context) CurrentUser {
	user, err := GetCurrentUser(c)
	if err != nil {
		panic(err)
	}
	return *user
}

func GetCurrentUserID(c *gin.Context) (uuid.UUID, error) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

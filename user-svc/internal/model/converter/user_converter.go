package converter

import (
	"go-saga-pattern/user-svc/internal/entity"
	"go-saga-pattern/user-svc/internal/model"
)

func UserToResponse(user *entity.User) *model.UserResponse {
	if user == nil {
		return nil
	}

	return &model.UserResponse{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}
}

package entity

import (
	"go-saga-pattern/commoner/constant/enum"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID           `db:"id"`
	Username string              `db:"username"`
	Password string              `db:"password"`
	Role     enum.RoomStatusEnum `db:"role"`
}

package enum

type RoomStatusEnum string

const (
	RoomStatusEnumActive   RoomStatusEnum = "ACTIVE"
	RoomStatusEnumClosed   RoomStatusEnum = "CLOSED"
	RoomStatusEnumArchived RoomStatusEnum = "ARCHIVED"
	RoomStatusEnumDeleted  RoomStatusEnum = "DELETED"
)

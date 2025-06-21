package enum

type LockTypeEnum string

const (
	LockTypeUpdateEnum LockTypeEnum = "FOR UPDATE"
	LockTypeShareEnum  LockTypeEnum = "FOR SHARE"
	LockTypeNoneEnum   LockTypeEnum = "NONE"
)

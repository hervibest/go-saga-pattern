package nullable

import "database/sql"

// ToSQLFloat64 mengonversi pointer float64 ke sql.NullFloat64
func ToSQLFloat64(input *float64) sql.NullFloat64 {
	if input == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{
		Float64: *input,
		Valid:   true,
	}
}

// ToSQLInt32 mengonversi pointer int32 ke sql.NullInt32
func ToSQLInt32(input *int32) sql.NullInt32 {
	if input == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{
		Int32: *input,
		Valid: true,
	}
}

// ToSQLInt64 mengonversi pointer int64 ke sql.NullInt64
func ToSQLInt64(input *int64) sql.NullInt64 {
	if input == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: *input,
		Valid: true,
	}
}

// ToSQLString mengonversi pointer string ke sql.NullString
func ToSQLString(input *string) sql.NullString {
	if input == nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *input,
		Valid:  true,
	}
}

func SQLStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func SQLFloat64ToPtr(ns sql.NullFloat64) *float64 {
	if ns.Valid {
		return &ns.Float64
	}
	return nil
}

func SQLtoString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func SQLtoFloat64(ns sql.NullFloat64) float64 {
	if ns.Valid {
		return ns.Float64
	}
	return 0
}

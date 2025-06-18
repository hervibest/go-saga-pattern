package nullable

import (
	"database/sql"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ToProtoDouble mengonversi sql.NullFloat64 ke *wrapperspb.DoubleValue
func SQLToProtoDouble(val sql.NullFloat64) *wrapperspb.DoubleValue {
	if !val.Valid {
		return nil
	}
	return wrapperspb.Double(val.Float64)
}

// ToProtoString mengonversi sql.NullString ke *wrapperspb.StringValue
func SQLToProtoString(val sql.NullString) *wrapperspb.StringValue {
	if !val.Valid {
		return nil
	}
	return wrapperspb.String(val.String)
}

func WrapDouble(v *wrapperspb.DoubleValue) *float64 {
	if v == nil {
		return nil
	}
	val := v.GetValue()
	return &val
}

func WrapString(v *wrapperspb.StringValue) *string {
	if v == nil {
		return nil
	}
	val := v.GetValue()
	return &val
}

package helper

import (
	"fmt"
	"go-saga-pattern/commoner/logs"
	"time"

	"go.uber.org/zap"
)

type TimeParserHelper interface {
	TimeParseInDefaultLocation(inputTime string) (time.Time, error)
	TimeParseRFC3339(input string) (time.Time, error)
	TimeFormatToDefaultLayout(t time.Time) string
	TimeFormatToRFC3339(t time.Time) string
	NowInDefaultLocation() time.Time
}

type timeParserHelper struct {
	log logs.Log
	loc *time.Location
}

func NewTimeParserHelper(log logs.Log) TimeParserHelper {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Error("failed to load location Asia/Jakarta", zap.Error(err))
		return nil
	}
	return &timeParserHelper{
		log: log,
		loc: loc,
	}
}

func (h *timeParserHelper) TimeParseInDefaultLocation(inputTime string) (time.Time, error) {
	parsedInputTime, err := time.ParseInLocation("2006-01-02 15:04:05", inputTime, h.loc)
	if err != nil {
		h.log.Error(fmt.Sprintf("Failed to parse local time with error : %v", err))
		return time.Time{}, err
	}
	return parsedInputTime, nil
}

func (h *timeParserHelper) TimeParseRFC3339(input string) (time.Time, error) {
	parsedTime, err := time.Parse(time.RFC3339, input)
	if err != nil {
		h.log.Error(fmt.Sprintf("Failed to parse RFC3339 time with error : %v", err))
		return time.Time{}, err
	}
	return parsedTime.In(h.loc), nil
}

func (h *timeParserHelper) TimeFormatToDefaultLayout(t time.Time) string {
	return t.In(h.loc).Format("2006-01-02 15:04:05")
}

func (h *timeParserHelper) TimeFormatToRFC3339(t time.Time) string {
	return t.In(h.loc).Format(time.RFC3339)
}

func (h *timeParserHelper) NowInDefaultLocation() time.Time {
	return time.Now().In(h.loc)
}

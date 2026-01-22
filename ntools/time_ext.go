package ntools

import (
	"time"

	"github.com/niexqc/nlibs/nerror"
)

func Time2Str(ptime time.Time) string {
	return ptime.Format("2006-01-02 15:04:05")
}

func Time2StrMilli(ptime time.Time) string {
	return ptime.Format("2006-01-02 15:04:05.000")
}

func TimeTo20060102(ptime time.Time) string {
	return ptime.Format("20060102")
}

func Time2StrByLayout(ptime time.Time, layout string) string {
	return ptime.Format(layout)
}

func TimeStr2Time(timeStr string) (time.Time, error) {
	return TimeStr2TimeByLayout(timeStr, "2006-01-02 15:04:05")
}

func TimeStr2TimeByLayout(timeStr, layout string) (time.Time, error) {
	timeObj, err := time.ParseInLocation(layout, timeStr, time.Local)
	if err != nil {
		return time.Time{}, nerror.NewRunTimeErrorWithError("时间解析错误", err)
	}
	return timeObj, nil
}

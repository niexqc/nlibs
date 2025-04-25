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

func TimeStr2Time(timeStr string) time.Time {
	time, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("时间解析错误", err))
	}
	return time
}

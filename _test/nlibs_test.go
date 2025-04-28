package test

import (
	"log/slog"
	"testing"

	"github.com/niexqc/nlibs/nerror"
)

func TestErrorExt(t *testing.T) {
	err := nerror.NewRunTimeError("this is run time error")
	slog.Info(nerror.GenErrDetail(err))

}

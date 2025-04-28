package test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
)

func TestErrorExt(t *testing.T) {
	err := nerror.NewRunTimeError("this is run time error")
	slog.Info(nerror.GenErrDetail(err))

}
func TestReadXlsx(t *testing.T) {
	contents, _ := ntools.XlsxRead("test.xlsx", "user", 1)
	fmt.Println(contents)
}

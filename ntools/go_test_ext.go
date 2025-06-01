package ntools

import (
	"fmt"
	"strings"
	"testing"

	"github.com/niexqc/nlibs/nerror"
)

func TestErrPanicMsg(t *testing.T, titleAndmsg string) {
	t.Error(titleAndmsg)
	panic(nerror.NewRunTimeError(titleAndmsg))
}

func TestErrPainic(t *testing.T, title string, err error) {
	if nil != err {
		errMsg := fmt.Sprintf(title+":异常:%v", err)
		t.Error(errMsg)
		panic(nerror.NewRunTimeErrorWithError(errMsg, err))
	}
}

func TestEq(t *testing.T, title string, exp, act any) {
	if exp != act {
		errMsg := fmt.Sprintf(title+":期望【%v】,实际:【%v】", exp, act)
		t.Error(errMsg)
		panic(nerror.NewRunTimeError(errMsg))
	}
}

func TestErrNotNil(t *testing.T, title string, err error) {
	if err == nil {
		errMsg := title + ":期望发生异常,但未发生"
		t.Error(errMsg)
		panic(nerror.NewRunTimeError(errMsg))
	}
}

func TestStrContains(t *testing.T, title string, expContains, act string) {
	if !strings.Contains(act, expContains) {
		errMsg := fmt.Sprintf(title+":期望包含【%v】,实际:【%v】", expContains, act)
		t.Error(errMsg)
		panic(nerror.NewRunTimeError(errMsg))
	}
}

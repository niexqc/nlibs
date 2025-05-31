package ntools

import (
	"strings"
	"testing"
)

func TestErrPanicMsg(t *testing.T, titleAndmsg string) {
	t.Error(titleAndmsg)
}

func TestErrPainic(t *testing.T, title string, err error) {
	if nil != err {
		t.Errorf(title+":异常:%v", err)
	}
}

func TestEq(t *testing.T, title string, exp, act any) {
	if exp != act {
		t.Errorf(title+":期望【%v】,实际:【%v】", exp, act)
	}
}

func TestStrContains(t *testing.T, title string, expContains, act string) {
	if !strings.Contains(act, expContains) {
		t.Errorf(title+":期望包含【%v】,实际:【%v】", expContains, act)
	}
}

package nerror

import (
	"fmt"
	"runtime/debug"
)

func GenErrDetail(err error) string {
	return fmt.Sprintf("%s\n%v", err.Error(), string(debug.Stack()))
}

func NewRunTimeError(errDesc string) *RunTimeErr {
	return &RunTimeErr{
		ErrDesc: errDesc,
	}
}

func NewRunTimeErrorWithError(errDesc string, err error) *RunTimeErr {
	return &RunTimeErr{
		ErrDesc: errDesc,
		SrcErr:  err,
	}
}

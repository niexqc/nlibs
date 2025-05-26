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

func NewRunTimeErrorFmt(fmtStr string, arg ...any) *RunTimeErr {
	return &RunTimeErr{
		ErrDesc: fmt.Sprintf(fmtStr, arg...),
	}
}

func NewRunTimeErrorWithError(errDesc string, err error) *RunTimeErr {
	return &RunTimeErr{
		ErrDesc: errDesc,
		SrcErr:  err,
	}
}

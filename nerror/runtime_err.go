package nerror

type RunTimeErr struct {
	ErrDesc string
	SrcErr  error
}

// Error 实现Error接口
func (e RunTimeErr) Error() string {
	if e.SrcErr != nil {
		return e.ErrDesc + "\n原始错误:" + e.SrcErr.Error()
	}
	return e.ErrDesc
}

package ntools

// 匿名函数
func If3[T any](cond bool, ok, no T) T {
	if cond {
		return ok
	} else {
		return no
	}
}

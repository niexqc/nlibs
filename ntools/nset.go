package ntools

type NSet[T comparable] map[T]struct{}

func NewNSet[T comparable]() NSet[T] {
	return make(NSet[T])
}

// 插入元素
func (s NSet[T]) Add(item T) {
	s[item] = struct{}{}
}

// 转换为切片
func (s NSet[T]) ToSlice() []T {
	slice := make([]T, 0, len(s))
	for item := range s {
		slice = append(slice, item)
	}
	return slice
}

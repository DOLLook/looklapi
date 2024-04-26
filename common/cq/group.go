package cq

// 泛型Group
type Group[E any] []*GroupEntry[E]

type GroupEntry[E any] struct {
	Key    any
	Values GenericSlice[E]
}

// 泛型切片分组
func (source GenericSlice[T]) GroupBy(f func(e T) any) Group[T] {
	var temp []*GroupEntry[T]

	if len(source) <= 0 || f == nil {
		return temp
	}

	m := make(map[any]*GroupEntry[T])

	for _, item := range source {
		k := f(item)
		if e := m[k]; e != nil {
			e.Values = append(e.Values, item)
		} else {
			m[k] = &GroupEntry[T]{
				Key:    k,
				Values: []T{item},
			}
		}
	}

	for _, v := range m {
		temp = append(temp, v)
	}

	return temp
}

// 分组迭代
func (source Group[T]) Foreach(f func(g *GroupEntry[T])) {
	if len(source) <= 0 || f == nil {
		return
	}

	for _, item := range source {
		f(item)
	}
}

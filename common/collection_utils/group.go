package collection_utils

// 泛型Group
type Group[E any] []*groupEntry[E]

type groupEntry[E any] struct {
	key    any
	Values GenericSlice[E]
}

// 泛型切片分组
func (source GenericSlice[T]) GroupBy(f func(item T) any) Group[T] {
	var temp []*groupEntry[T]

	if len(source) <= 0 || f == nil {
		return temp
	}

	m := make(map[any]*groupEntry[T])

	for _, item := range source {
		k := f(item)
		if e := m[k]; e != nil {
			e.Values = append(e.Values, item)
		} else {
			m[k] = &groupEntry[T]{
				key:    k,
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
func (source Group[T]) Foreach(f func(item *groupEntry[T])) {
	if len(source) <= 0 || f == nil {
		return
	}

	for _, item := range source {
		f(item)
	}
}

package cq

import (
	"golang.org/x/exp/maps"
	"slices"
)

// slice转map
func SliceToMap[S any, K comparable, V any](slice []S, kSelector func(e S) K, vSelector func(e S) V) GenericMap[K, V] {
	length := len(slice)
	if length < 1 || kSelector == nil || vSelector == nil {
		return nil
	}

	m := make(map[K]V, length)
	for _, e := range slice {
		k := kSelector(e)
		v := vSelector(e)
		m[k] = v
	}

	return m
}

// slice group后转map
func SliceGroupToMap[S any, K comparable, V any](slice []S, kSelector func(e S) K, vSelector func(g *GroupEntry[S]) V) GenericMap[K, V] {
	length := len(slice)
	if length < 1 || kSelector == nil || vSelector == nil {
		return nil
	}

	m := make(map[K]V, length)
	FromSlice(slice).GroupBy(func(e S) any {
		return kSelector(e)
	}).Foreach(func(g *GroupEntry[S]) {
		v := vSelector(g)
		m[g.Key.(K)] = v
	})

	return m
}

// 删除切片元素
// firstCount 删除前几个, 0全部删除
// 返回 删除后的新slice
func SliceRemove[S ~[]E, E comparable](s S, e E, firstCount int) S {
	if len(s) < 1 {
		return s
	}

	sLen := len(s)
	var indexChangeMap = make(map[int]int, sLen)
	//var removeIndex []int
	removeLen := 0

	if firstCount > 0 {
		// 仅删除前firstCount匹配项
		hasBreak := false
		//tempLen := len(removeIndex)
		j := 0
		for i := 0; i < sLen; i++ {
			if !hasBreak && s[i] == e {
				if removeLen < firstCount {
					//removeIndex = append(removeIndex, i-len(removeIndex))
					//tempLen = len(removeIndex)
					removeLen++
					if removeLen == firstCount {
						hasBreak = true
					}
					continue
				}
			}

			if j != i {
				indexChangeMap[j] = i
			}
			j++
		}
	} else {
		// 删除所有匹配项
		j := 0
		for i := 0; i < sLen; i++ {
			if s[i] == e {
				//removeIndex = append(removeIndex, i-len(removeIndex))
				removeLen++
				continue
			}

			if j != i {
				indexChangeMap[j] = i
			}
			j++
		}
	}

	//removeLen := len(removeIndex)
	if removeLen < 1 {
		return s
	}

	changeKeys := maps.Keys(indexChangeMap)
	slices.Sort(changeKeys)
	for _, key := range changeKeys {
		s[key] = s[indexChangeMap[key]]
	}

	newLen := sLen - removeLen
	s = s[:newLen]

	return s
}

// 删除满足条件f的切片元素
// firstCount 删除前几个, 0全部删除
// 返回 删除后的新slice
func SliceRemoveBy[S ~[]E, E any](s S, f func(e E) bool, firstCount int) S {
	if len(s) < 1 || f == nil {
		return s
	}

	sLen := len(s)
	var indexChangeMap = make(map[int]int, sLen)
	//var removeIndex []int
	removeLen := 0

	if firstCount > 0 {
		// 仅删除前firstCount匹配项
		hasBreak := false
		//tempLen := len(removeIndex)
		j := 0
		for i := 0; i < sLen; i++ {
			if !hasBreak && f((s)[i]) {
				if removeLen < firstCount {
					//removeIndex = append(removeIndex, i-len(removeIndex))
					//tempLen = len(removeIndex)
					removeLen++
					if removeLen == firstCount {
						hasBreak = true
					}

					continue
				}
			}

			if j != i {
				indexChangeMap[j] = i
			}
			j++
		}
	} else {
		// 删除所有匹配项
		j := 0
		for i := 0; i < sLen; i++ {
			if f((s)[i]) {
				//removeIndex = append(removeIndex, i-len(removeIndex))
				removeLen++
				continue
			}

			if j != i {
				indexChangeMap[j] = i
			}
			j++
		}
	}

	//removeLen := len(removeIndex)
	if removeLen < 1 {
		return s
	}

	changeKeys := maps.Keys(indexChangeMap)
	slices.Sort(changeKeys)
	for _, key := range changeKeys {
		s[key] = s[indexChangeMap[key]]
	}

	newLen := sLen - removeLen
	s = s[:newLen]

	return s
}

// 删除切片元素
// index 待删除的索引
// 返回 删除后的新slice
func SliceRemoveByIndex[S ~[]E, E any](s S, index ...int) S {
	sLen := len(s)
	if sLen < 1 {
		return s
	}

	if len(index) == 1 && index[0] < sLen {
		return slices.Delete(s, index[0], index[0]+1)
	}

	removeMapCheck := SliceToMap(index, func(e int) int {
		return e
	}, func(e int) bool {
		return false
	})

	change := false
	indexChangeMap := make(map[int]int, sLen)
	// 删除所有匹配项
	j := 0
	for i := 0; i < sLen; i++ {
		if hasRemove, ok := removeMapCheck[i]; ok {
			change = true
			if !hasRemove {
				removeMapCheck[i] = true
			}
			continue
		}

		if j != i {
			indexChangeMap[j] = i
		}
		j++
	}

	if !change {
		return s
	}

	changeKeys := maps.Keys(indexChangeMap)
	slices.Sort(changeKeys)
	for _, key := range changeKeys {
		s[key] = s[indexChangeMap[key]]
	}

	newLen := sLen - removeMapCheck.Count(nil, func(val bool) bool {
		return val
	})
	s = s[:newLen]

	return s
}

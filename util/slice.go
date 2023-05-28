package util

import (
	"math/rand"
	"time"

	"golang.org/x/exp/slices"
)

func Shuffle(in []int) {
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(in), func(i, j int) {
		in[i], in[j] = in[j], in[i]
	})
}

func Merge(arr1, arr2 [][]int) [][]int {
	if len(arr1) == 0 {
		return arr2
	}
	if len(arr2) == 0 {
		return arr1
	}
	result := make([][]int, 0)
	for _, outer := range arr1 {

		for _, inner := range arr2 {
			var res []int
			res = append(res, outer...)
			res = append(res, inner...)
			result = append(result, res)
		}
	}
	return result
}

func Contains(src []int, dst []int) bool {
	if len(src) == 0 && len(dst) == 0 {
		return true
	}

	for _, v := range dst {
		if !slices.Contains(src, v) {
			return false
		}
	}
	return true
}

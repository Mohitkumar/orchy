package util

import (
	"math/rand"
	"time"
)

func Shuffle(in []int) {
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(in), func(i, j int) {
		in[i], in[j] = in[j], in[i]
	})
}

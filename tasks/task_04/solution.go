package main

import "math"

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	n := len(nums)
	if n < 2 {
		return Stats{}
	}
	min_diff, max_diff := int64(math.MaxInt64), int64(math.MinInt64)

	for i := 1; i < n; i++ {
		diff := nums[i] - nums[i-1]
		if diff < min_diff {
			min_diff = diff
		}
		if diff > max_diff {
			max_diff = diff
		}
	}

	sum := nums[n-1] - nums[0]

	return Stats{
		Count: n - 1,
		Sum:   sum,
		Min:   min_diff,
		Max:   max_diff,
	}
}

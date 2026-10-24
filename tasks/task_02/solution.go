package main

func rotateRunes(s string, shift int) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	n := len(runes)
	ans := make([]rune, n)

	shift = (shift%n + n) % n
	for i := range runes {
		ans[i] = runes[(shift+i)%n]
	}
	return string(ans)
}

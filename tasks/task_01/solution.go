package main

import (
	"strings"
)

const defaultName = "World"

func greet(name string) string {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		trimmedName = defaultName
	}
	return "Hello, " + trimmedName + "!"
}

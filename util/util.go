package util

import (
	"strings"
)

var TabWidth = 8

func StringWidth(s string) int {
	return strings.Count(s, "") + strings.Count(s, "\t")*(TabWidth-1)
}

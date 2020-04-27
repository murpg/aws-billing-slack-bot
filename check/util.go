package main

import (
	"strconv"
	"strings"
)

func utilizeString(input *string) string {
	return strings.ToLower(strings.ReplaceAll(*input, " ", "-"))
}

func convertF(amount *string) float64 {
	r, _ := strconv.ParseFloat(*amount, 64)
	return r
}

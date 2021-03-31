package fts

import (
	"sort"
	"strings"
	"unicode"
)

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func analyze(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	tokens = stopwordFilter(tokens)
	return tokens
}

func mapToArr(m map[int]struct{}) []int {
	a := make([]int, len(m))
	i := -1
	for k := range m {
		i++
		a[i] = k
	}

	sort.Ints(a)

	return a
}

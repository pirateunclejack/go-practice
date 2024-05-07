package utils

import (
	"strings"

	"github.com/kljensen/snowball"
)

func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

func stopwordFilter(tokens []string) []string {
	var stopwords = map[string]struct{} {
		"a": {}, "and": {}, "be": {}, "have": {}, "i": {}, "in": {}, "of": {},
		"that": {}, "the": {}, "to": {}, 
	}

	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, ok := stopwords[token]; !ok {
			r = append(r, token)
		}
	}
	return r
}

func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	var err error
	for i, token := range tokens {
		r[i], err = snowball.Stem(token, "english", false)
		if err != nil {
			return nil
		}
	}
	return r
}

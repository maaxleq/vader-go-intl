package main

import (
	"strings"
	"sync"
)

// makeTokenBucket creates and fills a channel with tokens for rate limiting.
func makeTokenBucket(capacity int) chan struct{} {
	tokenBucket := make(chan struct{}, capacity)
	for i := 0; i < capacity; i++ {
		tokenBucket <- struct{}{}
	}

	return tokenBucket
}

// writeInTranslationMap safely writes a word and its corresponding value into a map.
func writeInTranslationMap[T any](mutex *sync.Mutex, dest map[string]T, word string, value T) {
	mutex.Lock()
	defer mutex.Unlock()

	dest[word] = value
}

// languageCodeToCamelCase converts a language code into a camel case format.
func languageCodeToCamelCase(lang string) string {
	return strings.ToUpper(string([]rune(lang)[:1])) + string([]rune(lang)[1:])
}

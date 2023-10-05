package main

import (
	"regexp"
	"sync"

	gt "github.com/bas24/googletranslatefree"
)

// translateMap translates a map's keys from English to another language.
func translateMap[T any](source map[string]T, lang string, tokenBucket chan struct{}) map[string]T {
	translatedMap := make(map[string]T)
	isWordRegex := regexp.MustCompile(`[0-9a-zA-Zà-üÀ-Ü\-\' ]+`)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for word, value := range source {
		if !isWordRegex.MatchString(word) {
			writeInTranslationMap[T](&mutex, translatedMap, word, value)
			continue
		}

		<-tokenBucket // Wait for token

		wg.Add(1)
		go func(value T, word, lang string) {
			defer func() {
				// Recover from potential panics in the Google Translate library.
				if r := recover(); r != nil {
					writeInTranslationMap[T](&mutex, translatedMap, word, value)
				}
				wg.Done()
			}()

			// Translate the word.
			translatedWord, errTranslate := gt.Translate(word, "en", lang)
			if errTranslate != nil {
				writeInTranslationMap[T](&mutex, translatedMap, word, value)
			} else {
				writeInTranslationMap[T](&mutex, translatedMap, translatedWord, value)
			}

			tokenBucket <- struct{}{} // Give back token
		}(value, word, lang)
	}

	// Wait for all translation goroutines to finish
	wg.Wait()

	return translatedMap
}

package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync/atomic"

	gt "github.com/bas24/googletranslatefree"
	"github.com/grassmudhorses/vader-go/lexicon"
)

type langSlice []string

func (s *langSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *langSlice) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

const maxRequestsAtOnce = 10

func makeTokenBucket(capacity int) chan struct{} {
	tokenBucket := make(chan struct{}, capacity)
	for i := 0; i < capacity; i++ {
		tokenBucket <- struct{}{}
	}

	return tokenBucket
}

func main() {
	var langs langSlice
	flag.Var(&langs, "langs", "language codes of the languages to add a Vader lexicon for")

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "whether to print progression")

	flag.Parse()

	isWordRegex := regexp.MustCompile(`[0-9a-zA-Zà-üÀ-Ü\-\' ]+`)

	tokenBucket := makeTokenBucket(maxRequestsAtOnce)

	for _, lang := range langs {
		var i uint32 = 0
		for word := range lexicon.Sentiments {
			if !isWordRegex.MatchString(word) {
				continue
			}

			<-tokenBucket // Wait for token

			go func(word, lang string) {
				defer func() {
					recover() // In case the Google Translate library panics, which happens sometimes
				}()

				translated, errTranslate := gt.Translate(word, "en", lang)
				if errTranslate != nil {
					log.Fatal(errTranslate)
				}

				if verbose {
					atomic.AddUint32(&i, 1)
					fmt.Printf("[%s:%d] %s = %s\n", lang, i, word, translated)
				}

				tokenBucket <- struct{}{} // Give back token
			}(word, lang)
		}
	}
}

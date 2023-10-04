package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/grassmudhorses/vader-go/lexicon"
	"github.com/mind1949/googletrans"
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

func main() {
	var langs langSlice
	flag.Var(&langs, "langs", "language codes of the languages to add a Vader lexicon for")

	flag.Parse()

	tokenBucket := make(chan struct{}, maxRequestsAtOnce)
	for i := 0; i < maxRequestsAtOnce; i++ {
		tokenBucket <- struct{}{}
	}

	for _, lang := range langs {
		for word := range lexicon.Sentiments {
			<-tokenBucket

			go func(word, lang string) {
				params := googletrans.TranslateParams{
					Src:  "en",
					Dest: lang,
					Text: word,
				}

				fmt.Printf("will translate %s to %s\n", word, lang)

				translated, errTranslate := googletrans.Translate(params)
				if errTranslate != nil {
					log.Fatal(errTranslate)
				}

				fmt.Printf("%s to %s = %s\n", word, lang, translated.Text)

				tokenBucket <- struct{}{}
			}(word, lang)
		}
	}
}

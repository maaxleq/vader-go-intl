package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/grassmudhorses/vader-go/lexicon"
)

// langSlice defines a type that can be used with the flag package to parse a comma-separated list.
type langSlice []string

func (s *langSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *langSlice) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

// Default number of concurrent requests for translations.
const defaultConcurrentRequests = 10

func main() {
	// Command-line flags initialization.
	var langs langSlice
	flag.Var(&langs, "langs", "language codes of the languages to add a Vader lexicon for (fr, es, nl, ...)")
	var concurentRequests int
	flag.IntVar(&concurentRequests, "reqs", defaultConcurrentRequests, "how many requests to run at once")
	var outPath string
	flag.StringVar(&outPath, "out", ".", "path where the translated lexicons folder will be created")
	flag.Parse()

	absOutPath, errOutPath := filepath.Abs(outPath)
	if errOutPath != nil {
		log.Fatal(errOutPath.Error())
	}

	tokenBucket := makeTokenBucket(concurentRequests)

	// For each specified language, translate the lexicons and write them to Go source files.
	for _, lang := range langs {
		sentiments := translateMap[float64](lexicon.Sentiments, lang, tokenBucket)
		contrasts := translateMap[bool](lexicon.Contrasts, lang, tokenBucket)
		negateList := translateMap[bool](lexicon.NegateList, lang, tokenBucket)
		boosters := translateMap[float64](lexicon.Boosters, lang, tokenBucket)

		if errWriteLexicon := writeTranslatedLexicon(absOutPath, lang, lexicon.CustomLexicon{
			Sentiments: sentiments,
			Contrasts:  contrasts,
			NegateList: negateList,
			Boosters:   boosters,
		}); errWriteLexicon != nil {
			log.Println(errWriteLexicon.Error())
		}
	}
}

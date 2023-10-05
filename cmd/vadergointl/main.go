package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

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

const defaultConcurrentRequests = 10

func makeTokenBucket(capacity int) chan struct{} {
	tokenBucket := make(chan struct{}, capacity)
	for i := 0; i < capacity; i++ {
		tokenBucket <- struct{}{}
	}

	return tokenBucket
}

func writeInTranslationMap[T any](mutex *sync.Mutex, dest map[string]T, word string, value T) {
	mutex.Lock()
	defer mutex.Unlock()

	dest[word] = value
}

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
				// In case the Google Translate library panics, which cna happen sometimes
				if r := recover(); r != nil {
					writeInTranslationMap[T](&mutex, translatedMap, word, value)
				}
				wg.Done()
			}()

			translatedWord, errTranslate := gt.Translate(word, "en", lang)
			if errTranslate != nil {
				writeInTranslationMap[T](&mutex, translatedMap, word, value)
			} else {
				writeInTranslationMap[T](&mutex, translatedMap, translatedWord, value)
			}

			tokenBucket <- struct{}{} // Give back token
		}(value, word, lang)
	}

	wg.Wait()

	return translatedMap
}

func languageCodeToCamelCase(lang string) string {
	return strings.ToUpper(string([]rune(lang)[:1])) + string([]rune(lang)[1:])
}

func writeTranslatedLexicon(outPath, lang string, lexicon lexicon.CustomLexicon) error {
	finalDir := path.Join(outPath, "lexicons")
	if _, errStat := os.Stat(finalDir); errStat != nil {
		if errMkdir := os.Mkdir(finalDir, 0666); errMkdir != nil {
			return errMkdir
		}
	}

	filename := fmt.Sprintf("lexicon_%s.go", lang)
	lexiconPackageName := "lexicons"
	lexiconStructName := fmt.Sprintf("Lexicon%s", languageCodeToCamelCase(lang))

	finalPath := path.Join(finalDir, filename)

	file, errFile := os.Create(finalPath)
	if errFile != nil {
		return errFile
	}

	return writeCode(file, lang, lexiconPackageName, lexiconStructName, lexicon)
}

func writeCode(file *os.File, lang, lexiconPackageName, lexiconStructName string, lexicon lexicon.CustomLexicon) error {
	defer file.Close()

	_, errWrite := fmt.Fprintf(file,
		`package %s

type %s struct {}

%s

%s

%s

%s

func (s %s) IsNegation(text string) bool {
	return negateList%s[text]
}

func (s %s) IsContrast(text string) bool {
	return contrasts%s[text]
}

func (s %s) Sentiment(text string) float64 {
	return sentiments%s[text]
}

func (s %s) BoostValue(text string) float64 {
	return boosters%s[text]
}
`,
		lexiconPackageName,
		lexiconStructName,
		makeMapCode[bool]("negateList"+languageCodeToCamelCase(lang), lexicon.NegateList),
		makeMapCode[bool]("contrasts"+languageCodeToCamelCase(lang), lexicon.Contrasts),
		makeMapCode[float64]("sentiments"+languageCodeToCamelCase(lang), lexicon.Sentiments),
		makeMapCode[float64]("boosters"+languageCodeToCamelCase(lang), lexicon.Boosters),
		lexiconStructName,
		languageCodeToCamelCase(lang),
		lexiconStructName,
		languageCodeToCamelCase(lang),
		lexiconStructName,
		languageCodeToCamelCase(lang),
		lexiconStructName,
		languageCodeToCamelCase(lang),
	)

	return errWrite
}

func makeMapCode[T any](varName string, source map[string]T) string {
	code := fmt.Sprintf(
		`var %s map[string]%T = map[string]%T{`, varName, *new(T), *new(T))

	for word, value := range source {
		escapedWord := strings.ReplaceAll(word, "`", "\u0060")
		code += fmt.Sprintf("\t`%s`: %v,\n", escapedWord, value)
	}

	code += "}"

	return code
}

func main() {
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

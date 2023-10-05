package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/grassmudhorses/vader-go/lexicon"
)

// writeTranslatedLexicon writes the translated lexicon into a Go source file.
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

	// Write the Go code to the file.
	return writeCode(file, lang, lexiconPackageName, lexiconStructName, lexicon)
}

// writeCode writes the code for the lexicon struct into a Go source file.
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

// makeMapCode generates Go code for a given map.
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

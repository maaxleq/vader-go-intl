// Package compose provides utilities to combine multiple lexicons for sentiment analysis
// into a single unified lexicon.
package compose

import "github.com/grassmudhorses/vader-go/lexicon"

type composedLexicon struct {
	innerLexicons []lexicon.Lexicon
}

// ComposeLexicons combines multiple lexicons into a single composed lexicon.
// Lookups on the composed lexicon will check each of the provided lexicons in turn.
func ComposeLexicons(lexicons ...lexicon.Lexicon) lexicon.Lexicon {
	return &composedLexicon{
		innerLexicons: lexicons,
	}
}

func (s *composedLexicon) IsNegation(text string) bool {
	for _, innerLexicon := range s.innerLexicons {
		if innerLexicon.IsNegation(text) {
			return true
		}
	}

	return false
}

func (s *composedLexicon) IsContrast(text string) bool {
	for _, innerLexicon := range s.innerLexicons {
		if innerLexicon.IsContrast(text) {
			return true
		}
	}

	return false
}

func (s *composedLexicon) Sentiment(text string) float64 {
	for _, innerLexicon := range s.innerLexicons {
		sentiment := innerLexicon.Sentiment(text)
		if sentiment != 0 {
			return sentiment
		}
	}

	return 0
}

func (s *composedLexicon) BoostValue(text string) float64 {
	for _, innerLexicon := range s.innerLexicons {
		boostValue := innerLexicon.BoostValue(text)
		if boostValue != 0 {
			return boostValue
		}
	}

	return 0
}

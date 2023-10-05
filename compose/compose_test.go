package compose_test

import (
	"testing"

	"github.com/maaxleq/vader-go-intl/compose"
)

type mockLexicon struct {
	negations   map[string]bool
	contrasts   map[string]bool
	sentiments  map[string]float64
	boostValues map[string]float64
}

func (m *mockLexicon) IsNegation(text string) bool {
	return m.negations[text]
}

func (m *mockLexicon) IsContrast(text string) bool {
	return m.contrasts[text]
}

func (m *mockLexicon) Sentiment(text string) float64 {
	return m.sentiments[text]
}

func (m *mockLexicon) BoostValue(text string) float64 {
	return m.boostValues[text]
}

func TestComposeLexicons(t *testing.T) {
	lexicon1 := &mockLexicon{
		negations:   map[string]bool{"not": true},
		contrasts:   map[string]bool{"but": true},
		sentiments:  map[string]float64{"good": 0.5},
		boostValues: map[string]float64{"very": 1.5},
	}

	lexicon2 := &mockLexicon{
		negations:   map[string]bool{"nope": true},
		contrasts:   map[string]bool{"although": true},
		sentiments:  map[string]float64{"bad": -0.5},
		boostValues: map[string]float64{"extremely": 2.0},
	}

	combined := compose.ComposeLexicons(lexicon1, lexicon2)

	tests := []struct {
		text      string
		isNeg     bool
		isContr   bool
		sentiment float64
		boostVal  float64
	}{
		{"not", true, false, 0, 0},
		{"but", false, true, 0, 0},
		{"good", false, false, 0.5, 0},
		{"very", false, false, 0, 1.5},
		{"nope", true, false, 0, 0},
		{"although", false, true, 0, 0},
		{"bad", false, false, -0.5, 0},
		{"extremely", false, false, 0, 2.0},
	}

	for _, test := range tests {
		if combined.IsNegation(test.text) != test.isNeg {
			t.Errorf("Expected %s IsNegation to be %v", test.text, test.isNeg)
		}
		if combined.IsContrast(test.text) != test.isContr {
			t.Errorf("Expected %s IsContrast to be %v", test.text, test.isContr)
		}
		if combined.Sentiment(test.text) != test.sentiment {
			t.Errorf("Expected %s Sentiment to be %v", test.text, test.sentiment)
		}
		if combined.BoostValue(test.text) != test.boostVal {
			t.Errorf("Expected %s BoostValue to be %v", test.text, test.boostVal)
		}
	}
}

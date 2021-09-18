package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

const topWordsCount = 10

var splitRegexp = regexp.MustCompile(
	// For splitting with '-': without chars or numbers around or then more than 1 repeat of '-'
	"(^|[-!\"#$%&'()*+,\\.:;<=>?@[\\]^_`{|}~\\s])+-($|[-!\"#$%&'()*+,\\.:;<=>?@[\\]^_`{|}~\\s])+|-{2,}" +
		// Other symbols to split
		"|[!\"#$%&'()*+,\\.:;<=>?@[\\]^_`{|}~\\s]+")

type wordFrequency struct {
	Word      string
	Frequency int
}

func Top10(s string) []string {
	words := countWordsFreq(splitRegexp.Split(s, -1))

	sort.Slice(words, func(i, j int) bool {
		if words[i].Frequency > words[j].Frequency {
			return true
		}
		if words[i].Frequency == words[j].Frequency {
			return cyrillicLess(words[i].Word, words[j].Word)
		}
		return false
	})

	res := make([]string, 0, topWordsCount)
	for i := 0; i < topWordsCount && i < len(words); i++ {
		res = append(res, words[i].Word)
	}
	return res
}

func countWordsFreq(words []string) []wordFrequency {
	indexes := make(map[string]int)
	frequencies := make([]wordFrequency, 0)
	for _, word := range words {
		if word == "" {
			continue
		}
		normWord := normalize(word)
		if index, ok := indexes[normWord]; ok {
			frequencies[index].Frequency++
			continue
		}
		frequencies = append(frequencies, wordFrequency{Word: normWord, Frequency: 1})
		indexes[normWord] = len(frequencies) - 1
	}
	return frequencies
}

func normalize(word string) string {
	return strings.ToLower(word)
}

// Treat "ё" as "e" for correct lexicographic sorting.
// Rune 'ё' (1105) is placed after 'я' (1103).
func cyrillicLess(a string, b string) bool {
	bRunes := []rune(b)
	for i, aRune := range a {
		if i >= len(bRunes) {
			return false
		}
		if aRune == 'ё' && bRunes[i] != 'ё' {
			return 'е' < bRunes[i]
		}
		if aRune != 'ё' && bRunes[i] == 'ё' {
			return aRune <= 'е'
		}
		if aRune != bRunes[i] {
			return aRune < bRunes[i]
		}
	}
	return true
}

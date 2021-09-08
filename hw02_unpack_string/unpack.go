package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

const (
	ignoredRune = rune(-1)
	escapeRune  = rune('\\')
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	runes := []rune(s)
	length := len(runes)
	if unicode.IsDigit(runes[0]) {
		return "", ErrInvalidString
	}

	prevRune := ignoredRune
	var escaped bool
	var err error
	b := strings.Builder{}
	for i := 0; i < length; i++ {
		if prevRune, escaped, err = processRune(runes[i], prevRune, escaped, &b); err != nil {
			return "", err
		}
	}
	if prevRune == escapeRune && !escaped {
		return "", ErrInvalidString
	}
	if !unicode.IsDigit(prevRune) || escaped {
		b.WriteRune(prevRune)
	}

	return b.String(), nil
}

func processRune(cur rune, prev rune, escaped bool, builder *strings.Builder) (rune, bool, error) {
	if unicode.IsDigit(prev) && !escaped && unicode.IsDigit(cur) {
		return cur, false, ErrInvalidString
	}
	if prev == escapeRune && (cur != escapeRune && !unicode.IsDigit(cur)) {
		return ignoredRune, false, ErrInvalidString
	}
	if !escaped && unicode.IsDigit(prev) {
		prev = ignoredRune
	}

	escaped = !escaped && prev == escapeRune
	if escaped && cur == escapeRune {
		prev = ignoredRune
	}

	if unicode.IsDigit(cur) {
		if escaped {
			return cur, escaped, nil
		}
		count, err := strconv.Atoi(string(cur))
		if err != nil {
			return rune(0), false, err
		}
		builder.WriteString(strings.Repeat(string(prev), count))
		return cur, escaped, nil
	}

	if prev != ignoredRune {
		builder.WriteRune(prev)
	}

	return cur, escaped, nil
}

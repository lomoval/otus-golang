package hw02unpackstring

import (
	"errors"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "d\n5abc", expected: "d\n\n\n\n\nabc"},
		// uncomment if task with asterisk completed
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},

		{input: string(rune(-1)), expected: "�"},
		{input: string(rune(-10)) + "2", expected: "��"},
		{input: string(rune(-100)) + `\2`, expected: "�2"},
		{input: "\u0000", expected: "\u0000"},
		{input: "\u00002", expected: "\u0000\u0000"},
		{input: "\u7777", expected: "\u7777"},
		{input: "\u77772", expected: "\u7777\u7777"},
		{input: `\\`, expected: `\`},
		{input: `\1`, expected: `1`},
		{input: `\2\3`, expected: `23`},
		{input: `\23a`, expected: `222a`},
		{input: `T2\2`, expected: `TT2`},
		{input: `#2!2$2`, expected: `##!!$$`},
		{input: `\\\\\\2`, expected: `\\\\`},
		{input: ` 2qwe\\\3`, expected: `  qwe\3`},
		{input: "  ", expected: "  "},
		{input: ` 1`, expected: ` `},
		{input: ` 2`, expected: `  `},
		{input: ` 0`, expected: ``},
		{input: ` 2 2 2`, expected: `      `},
		{input: `\1 2 2 2\1`, expected: `1      1`},
		{input: `qwe\3`, expected: `qwe3`},
		{input: `qwe\30`, expected: `qwe`},
		{input: `a0`, expected: ``},
		{input: `\00`, expected: ``},
		{input: `\10`, expected: ``},
		{input: `\0\0`, expected: `00`},
		{input: `a\0`, expected: `a0`},
		{input: `aa0`, expected: `a`},
		{input: `a1b1c1`, expected: `abc`},
		{input: `в1Ы1ф1`, expected: `вЫф`},
		{input: `-1`, expected: `-`},
		{input: `a-1`, expected: `a-`},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{`\0x000`, `qw\ne`, `0\0`, "3abc", "00", "45", "aaa10b", `\`, `\w`, `www\`}
	for _, tc := range invalidStrings {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}

func TestUnpackInvalidEscape(t *testing.T) {
	for i := 0; i < 255; i++ {
		if rune(i) == '\\' || unicode.IsDigit(rune(i)) {
			continue
		}
		_, err := Unpack(`\` + string(rune(i)))
		require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
	}
}

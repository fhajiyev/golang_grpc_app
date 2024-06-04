package utils

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Intersection func definition
func Intersection(str0s, str1s []string) []string {
	targetMap := make(map[string]bool, len(str0s))
	for _, str0 := range str0s {
		targetMap[str0] = true
	}

	result := make([]string, 0)

	for _, str1 := range str1s {
		if targetMap[str1] {
			result = append(result, str1)
		}
	}
	return result
}

// SliceItoA func definition
func SliceItoA(si []int) []string {
	sa := make([]string, 0, len(si))
	for _, a := range si {
		sa = append(sa, strconv.Itoa(a))
	}
	return sa
}

// SliceAtoi func definition
func SliceAtoi(sa []string) ([]int, error) {
	si := make([]int, 0, len(sa))
	for _, a := range sa {
		i, err := strconv.Atoi(a)
		if err != nil {
			return si, err
		}
		si = append(si, i)
	}
	return si, nil
}

type buffer struct {
	r         []byte
	runeBytes [utf8.UTFMax]byte
}

func (b *buffer) write(r rune) {
	if r < utf8.RuneSelf {
		b.r = append(b.r, byte(r))
		return
	}
	n := utf8.EncodeRune(b.runeBytes[0:], r)
	b.r = append(b.r, b.runeBytes[0:n]...)
}

func (b *buffer) indent() {
	if len(b.r) > 0 {
		b.r = append(b.r, '_')
	}
}

// CamelToUnderscore func definition
func CamelToUnderscore(s string) string {
	b := buffer{
		r: make([]byte, 0, len(s)),
	}
	var m rune
	var w bool
	for _, ch := range s {
		if unicode.IsUpper(ch) {
			if m != 0 {
				if !w {
					b.indent()
					w = true
				}
				b.write(m)
			}
			m = unicode.ToLower(ch)
		} else {
			if m != 0 {
				b.indent()
				b.write(m)
				m = 0
				w = false
			}
			b.write(ch)
		}
	}
	if m != 0 {
		if !w {
			b.indent()
		}
		b.write(m)
	}
	return string(b.r)
}

// UnderscoreToCamel changes underscore string into camel case string.
func UnderscoreToCamel(s string) string {
	titleString := strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)

	r := titleString[0]
	if r < unicode.MaxASCII && 'A' <= r && r <= 'Z' {
		r += 'a' - 'A'
	}

	return string(r) + titleString[1:]
}

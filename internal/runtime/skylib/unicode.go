package skylib

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

// Normalize normalizes a string
func UnicodeNormalize(s string, form string) string {
	switch form {
	case "NFC":
		return norm.NFC.String(s)
	case "NFD":
		return norm.NFD.String(s)
	case "NFKC":
		return norm.NFKC.String(s)
	case "NFKD":
		return norm.NFKD.String(s)
	default:
		return norm.NFC.String(s)
	}
}

// IsLetter checks if rune is a letter
func UnicodeIsLetter(r rune) bool {
	return unicode.IsLetter(r)
}

// IsDigit checks if rune is a digit
func UnicodeIsDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// IsSpace checks if rune is whitespace
func UnicodeIsSpace(r rune) bool {
	return unicode.IsSpace(r)
}

// IsUpper checks if rune is uppercase
func UnicodeIsUpper(r rune) bool {
	return unicode.IsUpper(r)
}

// IsLower checks if rune is lowercase
func UnicodeIsLower(r rune) bool {
	return unicode.IsLower(r)
}

// ToUpper converts to uppercase
func UnicodeToUpper(r rune) rune {
	return unicode.ToUpper(r)
}

// ToLower converts to lowercase
func UnicodeToLower(r rune) rune {
	return unicode.ToLower(r)
}

// Graphemes splits string into grapheme clusters
func UnicodeGraphemes(s string) []string {
	// Simplified implementation
	// Real implementation would use unicode/norm and proper boundary detection
	graphemes := []string{}

	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			break
		}
		graphemes = append(graphemes, s[:size])
		s = s[size:]
	}

	return graphemes
}

// Width returns display width of string (East Asian Width)
func UnicodeWidth(s string) int {
	totalWidth := 0

	for _, r := range s {
		prop := width.LookupRune(r)
		switch prop.Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			totalWidth += 2
		default:
			totalWidth += 1
		}
	}

	return totalWidth
}

// FoldCase performs case folding
func UnicodeFoldCase(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		runes[i] = unicode.SimpleFold(r)
	}
	return string(runes)
}

// Category returns Unicode category
func UnicodeCategory(r rune) string {
	if unicode.IsLetter(r) {
		return "Letter"
	}
	if unicode.IsDigit(r) {
		return "Digit"
	}
	if unicode.IsSpace(r) {
		return "Space"
	}
	if unicode.IsPunct(r) {
		return "Punct"
	}
	if unicode.IsSymbol(r) {
		return "Symbol"
	}
	return "Other"
}

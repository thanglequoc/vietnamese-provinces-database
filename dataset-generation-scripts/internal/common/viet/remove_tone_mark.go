package viet

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)
// RemoveVietToneMark removes Vietnamese tone marks and vowel diacritics from s,
// returning a Latin-only string (e.g. "Hà Nội" → "Ha Noi").
//
// It NFD-decomposes the input, strips all Unicode non-spacing marks (Mn),
// then NFC-recomposes. "đ"/"Đ" are replaced explicitly since they have no
// Unicode decomposition.
//
// Note: unicode.Mn removal affects all scripts, not only Vietnamese.
func RemoveVietToneMark(source string) string {
    trans := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
    result, _, err := transform.String(trans, source)
    if err != nil {
        // Should never happen for valid UTF-8 with these transforms.
        // Return source unchanged rather than silently returning a corrupt string.
        return source
    }
    result = strings.ReplaceAll(result, "đ", "d")
    result = strings.ReplaceAll(result, "Đ", "D")
    return result
}

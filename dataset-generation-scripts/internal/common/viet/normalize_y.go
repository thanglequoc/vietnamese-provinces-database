package viet

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormalizeIY normalizes the orthographic choice between "i" and "y" for the
// /i/ vowel following an initial consonant, as specified by Điều 9 of the
// Vietnamese orthography regulation.
//
// Rule (Điều 9.1): when the /i/ sound stands directly after an initial
// consonant with no medial glide and no final consonant, write it as "i",
// not "y".
//
//	"ký"  → "kí"
//	"kỷ"  → "kỉ"
//	"lý"  → "lí"
//	"mỹ"  → "mĩ"
//	"tỷ"  → "tỉ"    (tỷ lệ → tỉ lệ)
//
// Cases that are intentionally left unchanged:
//
//   - Syllables with no initial consonant: "ý", "yêu", "y tá" — here 'y' is
//     the nucleus in its own right and must not be changed.
//   - "uy" clusters after a consonant: "thuỷ", "nguỵ", "quy" — 'y' is the
//     nucleus after the medial glide 'u'; the rule does not apply.
//   - 'y' as a semivowel coda: "ngày", "tay", "dây" — 'y' closes the rhyme
//     and is not the standalone nucleus the rule targets.
//   - 'y' starting a diphthong: "yêu", "yến" (no initial consonant anyway).
//   - Proper nouns (Điều 9.2): "Vy", "Vỹ", "Thy" follow their own spelling.
//     This function cannot auto-detect proper nouns; callers are responsible
//     for excluding them.
//
// Like NormalizeToneMarks, the input is treated as space-separated tokens and
// each token is processed independently.

/* 
Note from repo author @thanglequoc: This function will just be here for reference and will not be used.
Because the administrative unit name is considered as their own proper noun, we will not apply this normalization to them. The function is kept here for reference in case we need to apply it to other fields in the future, but it will not be used for the administrative unit names.
*/
func NormalizeIY(s string) string {
	tokens := strings.Split(s, " ")
	for i, t := range tokens {
		tokens[i] = normalizeIYSyllable(t)
	}
	return strings.Join(tokens, " ")
}

// normalizeIYSyllable applies the Điều 9.1 rule to a single syllable token.
func normalizeIYSyllable(word string) string {
	if word == "" {
		return word
	}

	nfd := []rune(norm.NFD.String(word))
	gs := parseGraphemes(nfd)
	lower := baseLower(gs)

	// We need srcIdx for the gi-initial disambiguation in initialConsonantLen.
	// For this function there is no tone being moved, so srcIdx is irrelevant;
	// pass -1 (no tone source) — the gi special case only triggers on srcIdx==1,
	// which won't affect correctness here since we are not moving any tone mark.
	initLen := initialConsonantLen(lower, gs, -1)
	if initLen == 0 {
		// No initial consonant — Rule 9.1 does not apply.
		return word
	}

	codaLen := codaLen(lower, gs, initLen)

	vStart := initLen
	vEnd := len(gs) - codaLen

	// The rule targets exactly one vowel grapheme with base 'y', no coda.
	if vEnd-vStart != 1 || codaLen != 0 {
		return word
	}

	vowelG := gs[vStart]
	if vowelG.base != 'y' {
		return word
	}

	// Replace 'y'/'Y' with 'i'/'I', preserving case and all combining marks.
	if unicode.IsUpper(vowelG.orig) {
		gs[vStart].orig = 'I'
	} else {
		gs[vStart].orig = 'i'
	}
	gs[vStart].base = 'i'

	return norm.NFC.String(string(rebuildNFD(gs)))
}

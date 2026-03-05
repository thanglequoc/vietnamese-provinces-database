// Package viet provides Vietnamese tone mark normalization according to the
// standard convention described by scholar An Chi and encouraged by the
// Vietnamese education sector.
//
// Usage:
//
//	result := viet.NormalizeToneMarks("hoà bình")  // → "hoà bình"
//
// Dependency: golang.org/x/text (for Unicode normalization).
package viet

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ─── Public API ───────────────────────────────────────────────────────────────

// NormalizeToneMarks normalizes tone mark placement for every space-separated
// token in s according to Vietnamese orthographic convention.
func NormalizeToneMarks(s string) string {
	tokens := strings.Split(s, " ")
	for i, t := range tokens {
		tokens[i] = normalizeSyllable(t)
	}
	return strings.Join(tokens, " ")
}

// ─── Unicode / combining-mark helpers ────────────────────────────────────────

// Vietnamese tone combining marks (Unicode, NFD).
const (
	combGrave    = '\u0300' // huyền  (`)
	combAcute    = '\u0301' // sắc    (´)
	combHook     = '\u0309' // hỏi    (hook above)
	combTilde    = '\u0303' // ngã    (~)
	combDotBelow = '\u0323' // nặng   (dot below)
)

func isToneMark(r rune) bool {
	return r == combGrave || r == combAcute || r == combHook ||
		r == combTilde || r == combDotBelow
}

// Vietnamese base vowel letters (plain and pre-modified with horn/breve/hat).
var baseVowels = map[rune]bool{
	'a': true, 'ă': true, 'â': true,
	'e': true, 'ê': true,
	'i': true,
	'o': true, 'ô': true, 'ơ': true,
	'u': true, 'ư': true,
	'y': true,
}

func isVowel(r rune) bool { return baseVowels[unicode.ToLower(r)] }

// ─── Grapheme cluster representation (NFD) ───────────────────────────────────

type grapheme struct {
	orig  rune   // original rune (preserves case)
	base  rune   // lower-case base for comparison
	tones []rune // tone combining marks
	mods  []rune // non-tone combining marks (horn, breve, circumflex…)
}

func parseGraphemes(nfd []rune) []grapheme {
	var gs []grapheme
	for i := 0; i < len(nfd); {
		if unicode.Is(unicode.Mn, nfd[i]) {
			// Stray combining mark — attach to the previous grapheme.
			if len(gs) > 0 {
				last := &gs[len(gs)-1]
				if isToneMark(nfd[i]) {
					last.tones = append(last.tones, nfd[i])
				} else {
					last.mods = append(last.mods, nfd[i])
				}
			}
			i++
			continue
		}
		g := grapheme{orig: nfd[i], base: unicode.ToLower(nfd[i])}
		i++
		for i < len(nfd) && unicode.Is(unicode.Mn, nfd[i]) {
			if isToneMark(nfd[i]) {
				g.tones = append(g.tones, nfd[i])
			} else {
				g.mods = append(g.mods, nfd[i])
			}
			i++
		}
		gs = append(gs, g)
	}
	return gs
}

func rebuildNFD(gs []grapheme) []rune {
	var rs []rune
	for _, g := range gs {
		rs = append(rs, g.orig)
		rs = append(rs, g.mods...)
		rs = append(rs, g.tones...)
	}
	return rs
}

func baseLower(gs []grapheme) string {
	var sb strings.Builder
	for _, g := range gs {
		sb.WriteRune(g.base)
	}
	return sb.String()
}

// canonicalKey returns a string key for a grapheme that encodes both its base
// letter and its non-tone modifiers, so that ê (e+circumflex), ơ (o+horn),
// ư (u+horn), ô (o+circumflex), ă (a+breve) are all distinguishable from
// their plain counterparts.
//
// We map modifier combinations to a single sentinel rune:
//
//	e + circumflex (0302)  → 'ê'
//	o + circumflex (0302)  → 'ô'
//	a + circumflex (0302)  → 'â'
//	o + horn       (031b)  → 'ơ'
//	u + horn       (031b)  → 'ư'
//	a + breve      (0306)  → 'ă'
func canonicalKey(g grapheme) rune {
	const (
		circumflex = '\u0302'
		horn       = '\u031b'
		breve      = '\u0306'
	)
	hasMod := func(m rune) bool {
		for _, x := range g.mods {
			if x == m {
				return true
			}
		}
		return false
	}
	switch g.base {
	case 'e':
		if hasMod(circumflex) {
			return 'ê'
		}
	case 'o':
		if hasMod(circumflex) {
			return 'ô'
		}
		if hasMod(horn) {
			return 'ơ'
		}
	case 'u':
		if hasMod(horn) {
			return 'ư'
		}
	case 'a':
		if hasMod(circumflex) {
			return 'â'
		}
		if hasMod(breve) {
			return 'ă'
		}
	}
	return g.base
}

// canonicalLower returns a string built from canonicalKey for each grapheme,
// suitable for diphthong/cluster prefix matching.
func canonicalLower(gs []grapheme) string {
	var sb strings.Builder
	for _, g := range gs {
		sb.WriteRune(canonicalKey(g))
	}
	return sb.String()
}

// ─── Multi-character initial consonant clusters ───────────────────────────────

// Ordered longest-first so prefix matching is greedy.
var multiCharInitials = []string{
	"ngh", "gh",
	"gi", "qu",
	"ch", "kh", "ng", "nh", "ph", "th", "tr",
}

// finalConsonants lists letters that can appear as a single coda consonant.
// Includes 'k' which represents the same coda sound as 'c' and appears in
// ethnic minority place names (e.g. Ngoḳ, Buḱ, Krông Buḱ).
var finalConsonants = map[rune]bool{
	'c': true, 'h': true, 'k': true, 'm': true, 'n': true, 'p': true, 't': true,
}

// finalDigraphs lists two-letter coda clusters.
var finalDigraphs = []string{"ng", "nh", "ch"}

// ─── Core normalization ───────────────────────────────────────────────────────

func normalizeSyllable(word string) string {
	if word == "" {
		return word
	}

	nfd := []rune(norm.NFD.String(word))
	gs := parseGraphemes(nfd)

	// Collect all tone marks and note the source grapheme.
	var tones []rune
	srcIdx := -1
	for i, g := range gs {
		if len(g.tones) > 0 {
			tones = append(tones, g.tones...)
			srcIdx = i
		}
	}
	if len(tones) == 0 || srcIdx < 0 {
		return word // level tone (ngang) — nothing to move
	}

	// Special case: tone mark is attached to a non-letter base (e.g. apostrophe
	// in "M'́Drăk"). Strip it from the non-letter grapheme so toneTarget can
	// find the correct vowel target without that grapheme polluting srcIdx.
	if !isVowel(gs[srcIdx].base) && !isConsonantBase(gs[srcIdx].base) {
		gs[srcIdx].tones = nil
		// Re-run from scratch with srcIdx pointing to -1; toneTarget will
		// scan all vowels and use the first one as the target.
		srcIdx = -1
	} else {
		// Detach tone from current position.
		gs[srcIdx].tones = nil
	}

	targetIdx := toneTarget(gs, srcIdx)
	if targetIdx < 0 {
		// Cannot determine placement — restore and return original.
		if srcIdx >= 0 {
			gs[srcIdx].tones = tones
		}
		return word
	}

	gs[targetIdx].tones = tones
	return norm.NFC.String(string(rebuildNFD(gs)))
}

// isConsonantBase reports whether r is a base letter that can serve as a
// Vietnamese consonant (i.e. a letter that is not a vowel).
func isConsonantBase(r rune) bool {
	if r == 0 {
		return false
	}
	return unicode.IsLetter(r) && !isVowel(r)
}

// toneTarget returns the index into gs of the grapheme that should carry
// the tone mark for this syllable, applying rules §II – §VI.
// srcIdx is the grapheme index where the tone mark originally sat (before detaching).
func toneTarget(gs []grapheme, srcIdx int) int {
	// baseLower uses plain base letters — good for consonant identification.
	lower := baseLower(gs)

	initLen := initialConsonantLen(lower, gs, srcIdx)
	cLen := codaLen(lower, gs, initLen)

	vStart := initLen      // first vowel grapheme index
	vEnd := len(gs) - cLen // exclusive

	if vStart >= vEnd {
		return -1
	}

	hasCoda := cLen > 0

	// Skip any leading non-vowel graphemes inside the vowel span.
	// This handles ethnic minority place names with consonant clusters like
	// "kr", "hr", "kn" where initialConsonantLen only strips the first consonant,
	// leaving the rest (r, n, etc.) at the start of the vowel span.
	// Example: "Kŕai" → initLen=1(k), vStart=1, gs[1].base='r' (not a vowel)
	//          → advance vStart to 2 (the actual vowel 'a').
	for vStart < vEnd && !isVowel(gs[vStart].base) {
		vStart++
	}
	if vStart >= vEnd {
		return -1
	}

	// canonicalLower encodes modifier diacritics (ê≠e, ô≠o, ơ≠o, ư≠u, etc.)
	// so diphthong prefix matching works correctly.
	vLower := canonicalLower(gs[vStart:vEnd])

	abs := func(rel int) int { return vStart + rel }

	vRunes := []rune(vLower)

	// Detect whether the vowel span begins with a medial glide [w].
	//
	// A medial glide is plain 'u' or 'o' (no horn/circumflex) before another vowel,
	// but only when it genuinely acts as a glide, not as the nucleus itself.
	//
	// Two exceptions where the leading vowel is NOT a medial:
	//
	//  1. 'u' + 'a' without coda → "ua" open diphthong (Rule 4b), tone on 'u'.
	//
	//  2. 'u' + 'y' with NO initial consonant (initLen==0) → "ủy"-type syllable.
	//     Here 'u' is the nucleus and 'y' is a glide coda, so Rule 2 does not
	//     apply. Examples: ủy, ủy ban — tone stays on 'u'.
	//     Contrast: thuỷ, nguỵ — onset present → 'u' IS the medial → tone on 'y'.
	//
	// 'o' + 'a'/'e' always triggers medial (§VI applies with or without onset).
	medialOffset := 0
	if len(vRunes) >= 2 && isMedialGlide(vRunes[0]) {
		isUAOpenDiphthong := vRunes[0] == 'u' && vRunes[1] == 'a' && !hasCoda
		isUYNoOnset := vRunes[0] == 'u' && vRunes[1] == 'y' && initLen == 0
		if !isUAOpenDiphthong && !isUYNoOnset {
			medialOffset = 1
		}
	}

	// Remaining vowel string after any medial glide.
	vAfterMedial := string(vRunes[medialOffset:])

	// ── Rule 3a (§IV-1)
	// Diphthong iê/yê/uô/ươ WITH coda → tone on the 2nd letter of the diphthong.
	// This applies whether or not there is a preceding medial glide:
	//   tiến  → init=t,  vowels=[i,ê],   coda=n  → medialOffset=0, diphthong at 0 → abs(1)
	//   chuyến→ init=ch, vowels=[u,y,ê], coda=n  → medialOffset=1, diphthong at 1 → abs(2)
	if hasCoda {
		for _, dp := range []string{"iê", "yê", "uô", "ươ"} {
			if strings.HasPrefix(vAfterMedial, dp) {
				return abs(medialOffset + 1) // 2nd letter of diphthong
			}
		}
	}

	// ── Rule 3b (§IV-2 / §V)
	// Diphthong ia/ya/ua/ưa WITHOUT coda → tone on the 1st letter of the diphthong,
	// except for the gi-/qu- special cases.
	if !hasCoda {
		switch {
		case strings.HasPrefix(vAfterMedial, "ia"):
			// §V-1: "gi" initial → tone on 'a'; otherwise → 'i'
			if initLen >= 2 && strings.HasPrefix(lower, "gi") {
				return abs(medialOffset + 1)
			}
			return abs(medialOffset)

		case strings.HasPrefix(vAfterMedial, "ya"):
			return abs(medialOffset)

		case strings.HasPrefix(vAfterMedial, "ua"):
			// §V-2: "qu" initial → tone on 'a'; otherwise → 'u'
			if initLen >= 2 && strings.HasPrefix(lower, "qu") {
				return abs(medialOffset + 1)
			}
			return abs(medialOffset)

		case strings.HasPrefix(vAfterMedial, "ưa"):
			return abs(medialOffset)
		}
	}

	// ── Rule 2 / §VI
	// Medial glide clusters "oa", "oe", "uy" (and with coda: "oàn", "uyển", etc.)
	// → tone on the 2nd letter (the main vowel), i.e. medialOffset=1, target=abs(1).
	// This is already handled by medialOffset logic below, but we also need to
	// explicitly catch the 3-letter vowel spans like "uye" (in "uyển") where
	// medialOffset=1 and vAfterMedial starts with a single vowel.
	if medialOffset == 1 {
		// Single nucleus after medial glide (Rules 2/§III): tone on nucleus.
		return abs(1)
	}

	// ── Rules 1 & 2 fallback: single vowel, no medial glide.
	return abs(0)
}

// isMedialGlide reports whether r is a medial glide letter ('u' or 'o').
func isMedialGlide(r rune) bool { return r == 'u' || r == 'o' }

// ─── Structural analysis helpers ─────────────────────────────────────────────

// initialConsonantLen returns the number of graphemes forming the initial consonant.
// srcIdx is where the tone mark originally sat; used to detect the "gịa" special case.
func initialConsonantLen(lower string, gs []grapheme, srcIdx int) int {
	for _, mc := range multiCharInitials {
		if strings.HasPrefix(lower, mc) {
			n := len([]rune(mc))
			if mc == "gi" {
				// "gi" is an initial cluster only when followed by a vowel.
				if n < len(gs) && isVowel(gs[n].base) {
					// Special case: if the tone was originally on 'i' (index 1),
					// then 'i' is a vowel (part of "ia" diphthong), not the initial.
					// This handles "gịa" (giặt gịa) where g=initial, ia=diphthong.
					if srcIdx == 1 {
						return 1 // only 'g' is the initial
					}
					return n // full "gi" is the initial
				}
				// Otherwise fall through to single-letter initial 'g'.
				break
			}
			return n
		}
	}
	if len(gs) > 0 && !isVowel(gs[0].base) {
		return 1
	}
	return 0
}

// codaLen returns the number of trailing graphemes forming the coda.
func codaLen(lower string, gs []grapheme, initLen int) int {
	n := len(gs)
	if n-initLen <= 1 {
		return 0 // only one character after initial — no room for coda
	}

	// Two-letter digraph codas.
	if len(lower) >= 2 {
		last2 := lower[len(lower)-2:]
		for _, d := range finalDigraphs {
			if last2 == d {
				preIdx := n - 3
				if preIdx >= initLen && isVowel(gs[preIdx].base) {
					return 2
				}
			}
		}
	}

	// Single consonant coda.
	last := gs[n-1]
	if finalConsonants[last.base] && n-1 > initLen && isVowel(gs[n-2].base) {
		return 1
	}

	// Semivowel coda ('i', 'u', 'o' after a different vowel).
	if (last.base == 'i' || last.base == 'u' || last.base == 'o') && n-1 > initLen {
		prev := gs[n-2]
		if isVowel(prev.base) && prev.base != last.base {
			return 1
		}
	}

	return 0
}
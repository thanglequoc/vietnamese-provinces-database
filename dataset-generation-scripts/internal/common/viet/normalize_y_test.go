package viet

import "testing"

var normalizeIYCases = []struct {
	name  string
	input string
	want  string
}{
	// ════════════════════════════════════════════════════════════════════════
	// Core rule: C + y (no medial, no coda) → C + i
	// All 5 tones + ngang on each
	// ════════════════════════════════════════════════════════════════════════

	// ky → ki
	{"ky ngang", "ky", "ki"},
	{"kỳ huyền", "kỳ", "kì"},
	{"ký sắc", "ký", "kí"},
	{"kỷ hỏi", "kỷ", "kỉ"},
	{"kỹ ngã", "kỹ", "kĩ"},
	{"kỵ nặng", "kỵ", "kị"},

	// ly → li
	{"ly ngang", "ly", "li"},
	{"lỳ huyền", "lỳ", "lì"},
	{"lý sắc", "lý", "lí"},
	{"lỷ hỏi", "lỷ", "lỉ"},
	{"lỹ ngã", "lỹ", "lĩ"},
	{"lỵ nặng", "lỵ", "lị"},

	// my → mi
	{"my ngang", "my", "mi"},
	{"mỳ huyền", "mỳ", "mì"},
	{"mý sắc", "mý", "mí"},
	{"mỷ hỏi", "mỷ", "mỉ"},
	{"mỹ ngã", "mỹ", "mĩ"},
	{"mỵ nặng", "mỵ", "mị"},

	// ty → ti
	{"ty ngang", "ty", "ti"},
	{"tỳ huyền", "tỳ", "tì"},
	{"tý sắc", "tý", "tí"},
	{"tỷ hỏi", "tỷ", "tỉ"},
	{"tỹ ngã", "tỹ", "tĩ"},
	{"tỵ nặng", "tỵ", "tị"},

	// sy → si
	{"sy ngang", "sy", "si"},
	{"sỳ", "sỳ", "sì"},
	{"sý", "sý", "sí"},
	{"sỷ", "sỷ", "sỉ"},
	{"sỹ", "sỹ", "sĩ"},
	{"sỵ", "sỵ", "sị"},

	// hy → hi
	{"hy ngang", "hy", "hi"},
	{"hỳ", "hỳ", "hì"},
	{"hý", "hý", "hí"},
	{"hỷ", "hỷ", "hỉ"},
	{"hỹ", "hỹ", "hĩ"},
	{"hỵ", "hỵ", "hị"},

	// vy → vi
	{"vy ngang", "vy", "vi"},
	{"vỳ", "vỳ", "vì"},
	{"vý", "vý", "ví"},
	{"vỷ", "vỷ", "vỉ"},
	{"vỹ", "vỹ", "vĩ"},
	{"vỵ", "vỵ", "vị"},

	// by → bi
	{"by ngang", "by", "bi"},
	{"bý", "bý", "bí"},

	// dy → di (đ already covered by RemoveVietToneMark, but d+y is valid)
	{"dy ngang", "dy", "di"},
	{"dý", "dý", "dí"},

	// ny → ni
	{"ny ngang", "ny", "ni"},
	{"ný", "ný", "ní"},

	// ════════════════════════════════════════════════════════════════════════
	// Multi-letter initials
	// ════════════════════════════════════════════════════════════════════════
	{"thy → thi", "thy", "thi"},
	{"thỳ", "thỳ", "thì"},
	{"thý", "thý", "thí"},
	{"thỷ", "thỷ", "thỉ"},
	{"thỹ", "thỹ", "thĩ"},
	{"thỵ", "thỵ", "thị"},

	{"nhy → nhi", "nhy", "nhi"},
	{"nhý", "nhý", "nhí"},
	{"nhỷ", "nhỷ", "nhỉ"},

	{"khy → khi", "khy", "khi"},
	{"khy ngang", "khy", "khi"},

	{"phy → phi", "phy", "phi"},
	{"phý", "phý", "phí"},

	{"try → tri", "try", "tri"},
	{"trý", "trý", "trí"},

	{"ngy → ngi (rare)", "ngy", "ngi"},

	// ════════════════════════════════════════════════════════════════════════
	// Real words from the regulation examples
	// ════════════════════════════════════════════════════════════════════════
	{"kỉ niệm", "kỷ", "kỉ"},
	{"lí luận", "lý", "lí"},
	{"mĩ thuật", "mỹ", "mĩ"},
	{"tỉ lệ", "tỷ", "tỉ"},
	{"bác sĩ", "sỷ", "sỉ"},

	// ════════════════════════════════════════════════════════════════════════
	// Uppercase / mixed case — 'Y' → 'I', case preserved
	// ════════════════════════════════════════════════════════════════════════
	{"KY uppercase", "KY", "KI"},
	{"Ký mixed", "Ký", "Kí"},
	{"LÝ", "LÝ", "LÍ"},
	{"Mỹ", "Mỹ", "Mĩ"},

	// ════════════════════════════════════════════════════════════════════════
	// Must NOT be changed — no initial consonant
	// ════════════════════════════════════════════════════════════════════════
	{"y standalone ngang", "y", "y"},
	{"ý no initial", "ý", "ý"},
	{"ỳ no initial", "ỳ", "ỳ"},
	{"ỷ no initial", "ỷ", "ỷ"},
	{"ỹ no initial", "ỹ", "ỹ"},
	{"ỵ no initial", "ỵ", "ỵ"},

	// ════════════════════════════════════════════════════════════════════════
	// Must NOT be changed — 'y' as semivowel coda (after another vowel)
	// ════════════════════════════════════════════════════════════════════════
	{"ngày coda y", "ngày", "ngày"},
	{"tay coda y", "tay", "tay"},
	{"dây coda y", "dây", "dây"},
	{"mày coda y", "mày", "mày"},
	{"xây coda y", "xây", "xây"},

	// ════════════════════════════════════════════════════════════════════════
	// Must NOT be changed — 'y' in "uy" cluster after consonant
	// (u = medial glide, y = nucleus; Điều 9 does not apply)
	// ════════════════════════════════════════════════════════════════════════
	{"thuỷ uy cluster", "thuỷ", "thuỷ"},
	{"nguỵ uy cluster", "nguỵ", "nguỵ"},
	{"buýt uy+coda", "buýt", "buýt"},
	{"suýt uy+coda", "suýt", "suýt"},
	{"quy qu-initial", "quy", "qui"},   // qu treated as initial → y alone → normalize
	{"quý qu-initial", "quý", "quí"},   // same

	// ════════════════════════════════════════════════════════════════════════
	// Must NOT be changed — 'y' starting a diphthong yê
	// ════════════════════════════════════════════════════════════════════════
	{"yêu no initial", "yêu", "yêu"},
	{"yến no initial", "yến", "yến"},

	// ════════════════════════════════════════════════════════════════════════
	// Multi-word sentences
	// ════════════════════════════════════════════════════════════════════════
	{"hi vọng", "hi vọng", "hi vọng"},       // already correct
	{"hy vọng → hi vọng", "hy vọng", "hi vọng"},
	{"kỷ niệm → kỉ niệm", "kỷ niệm", "kỉ niệm"},
	{"lý luận → lí luận", "lý luận", "lí luận"},
	{"mỹ thuật → mĩ thuật", "mỹ thuật", "mĩ thuật"},
	{"tỷ lệ → tỉ lệ", "tỷ lệ", "tỉ lệ"},
	{"bác sỹ → bác sĩ", "bác sỹ", "bác sĩ"},
	{"mixed sentence", "lý thuyết kỹ thuật", "lí thuyết kĩ thuật"},
	// Proper noun 'Vy' — caller responsibility; function normalizes it
	// (document that proper nouns must be excluded by the caller)
}

func TestNormalizeIY(t *testing.T) {
	for _, tc := range normalizeIYCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeIY(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeIY(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

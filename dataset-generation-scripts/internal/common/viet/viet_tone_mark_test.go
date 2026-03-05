package viet

import "testing"

var testCases = []struct {
	name  string
	input string
	want  string
}{
	// ════════════════════════════════════════════════════════════════════════
	// Rule 1 — Non-rounded syllable, single vowel: tone on the main vowel
	// ════════════════════════════════════════════════════════════════════════
	{"r1 á", "á", "á"},
	{"r1 tã", "tã", "tã"},
	{"r1 nhà", "nhà", "nhà"},
	{"r1 nhãn", "nhãn", "nhãn"},
	{"r1 gánh", "gánh", "gánh"},
	{"r1 ngáng", "ngáng", "ngáng"},
	// All 5 tones on a simple CVC syllable
	{"r1 bàn", "bàn", "bàn"},
	{"r1 bán", "bán", "bán"},
	{"r1 bản", "bản", "bản"},
	{"r1 bãn", "bãn", "bãn"},
	{"r1 bạn", "bạn", "bạn"},
	// Modified vowels (â, ă, ê, ô, ơ, ư) — single nucleus
	{"r1 thân", "thân", "thân"},
	{"r1 lăn", "lăn", "lăn"},
	{"r1 kền", "kền", "kền"},
	{"r1 tốt", "tốt", "tốt"},
	{"r1 lớn", "lớn", "lớn"},
	// gh- initial
	{"r1 ghé", "ghé", "ghé"},
	{"r1 ghẹ", "ghẹ", "ghẹ"},
	{"r1 ghì", "ghì", "ghì"},

	// ════════════════════════════════════════════════════════════════════════
	// Rule 2 — Rounded syllable (medial [w]), single vowel: tone on nucleus
	// ════════════════════════════════════════════════════════════════════════
	{"r2 hoà", "hoà", "hoà"},
	{"r2 hoè", "hoè", "hoè"},
	{"r2 quỳ", "quỳ", "quỳ"},
	{"r2 quà", "quà", "quà"},
	{"r2 quờ", "quờ", "quờ"},
	{"r2 thuỷ", "thuỷ", "thuỷ"},
	{"r2 nguỵ", "nguỵ", "nguỵ"},
	{"r2 hoàn", "hoàn", "hoàn"},
	{"r2 quét", "quét", "quét"},
	{"r2 quát", "quát", "quát"},
	{"r2 quỵt", "quỵt", "quỵt"},
	{"r2 suýt", "suýt", "suýt"},
	// uy with coda (xe buýt)
	{"r2 buýt", "buýt", "buýt"},
	{"r2 tuýt", "tuýt", "tuýt"},
	// oe with coda
	{"r2 khoẻ", "khoẻ", "khoẻ"},
	// oa/oe/uy — standalone (no initial)
	{"r2 oà", "oà", "oà"},
	{"r2 oè", "oè", "oè"},
	// uy standalone (no initial) — u is nucleus, tone moves to u
	{"r2 uỳ standalone", "uỳ", "ùy"},
	// ng- initial + oa
	{"r2 ngoài", "ngoài", "ngoài"},
	{"r2 ngoạn", "ngoạn", "ngoạn"},
	// All 5 tones on "hoXn"
	{"r2 hoàn huyền", "hoàn", "hoàn"},
	{"r2 hoán sắc", "hoán", "hoán"},
	{"r2 hoản hỏi", "hoản", "hoản"},
	{"r2 hoãn ngã", "hoãn", "hoãn"},
	{"r2 hoạn nặng", "hoạn", "hoạn"},

	// ════════════════════════════════════════════════════════════════════════
	// Rule 3a — Diphthong iê/yê/uô/ươ WITH coda: tone on 2nd letter
	// ════════════════════════════════════════════════════════════════════════
	{"r3a yếu", "yếu", "yếu"},
	{"r3a uốn", "uốn", "uốn"},
	{"r3a ườn", "ườn", "ườn"},
	{"r3a tiến", "tiến", "tiến"},
	{"r3a muốn", "muốn", "muốn"},
	{"r3a mượn", "mượn", "mượn"},
	{"r3a thiện", "thiện", "thiện"},
	{"r3a người", "người", "người"},
	{"r3a viếng", "viếng", "viếng"},
	{"r3a muống", "muống", "muống"},
	{"r3a cường", "cường", "cường"},
	// iê + various codas
	{"r3a tiền", "tiền", "tiền"},
	{"r3a tiếp", "tiếp", "tiếp"},
	{"r3a tiệc", "tiệc", "tiệc"},
	{"r3a tiếng", "tiếng", "tiếng"},
	{"r3a tiệm", "tiệm", "tiệm"},
	// iê + semivowel coda (u/o)
	{"r3a điều", "điều", "điều"},
	{"r3a chiều", "chiều", "chiều"},
	{"r3a nhiều", "nhiều", "nhiều"},
	// ươ + various codas
	{"r3a bướm", "bướm", "bướm"},
	{"r3a tướng", "tướng", "tướng"},
	{"r3a đường", "đường", "đường"},
	// ươ + semivowel coda
	{"r3a tươi", "tươi", "tươi"},
	// uô + semivowel coda
	{"r3a nuối", "nuối", "nuối"},
	{"r3a muội", "muội", "muội"},
	{"r3a buổi", "buổi", "buổi"},
	// ngh- initial (3-letter cluster) + iê
	{"r3a nghiêng", "nghiêng", "nghiêng"},
	{"r3a nghiện", "nghiện", "nghiện"},
	// medial [w] + diphthong yê + coda (the chuyến-family)
	{"r3a medial chuyến", "chuyến", "chuyến"},
	{"r3a thuyền", "thuyền", "thuyền"},
	{"r3a khuyết", "khuyết", "khuyết"},
	{"r3a luyện", "luyện", "luyện"},
	{"r3a tuyến", "tuyến", "tuyến"},
	{"r3a huyền", "huyền", "huyền"},
	// All 5 tones on "chuyXn"
	{"r3a chuyền", "chuyền", "chuyền"},
	{"r3a chuyến", "chuyến", "chuyến"},
	{"r3a chuyển", "chuyển", "chuyển"},
	{"r3a chuyễn", "chuyễn", "chuyễn"},
	{"r3a chuyện", "chuyện", "chuyện"},
	// All 5 tones on "ườn/ướn/ưởn/ưỡn/ượn"
	{"r3a ườn", "ườn", "ườn"},
	{"r3a ướn", "ướn", "ướn"},
	{"r3a ưởn", "ưởn", "ưởn"},
	{"r3a ưỡn", "ưỡn", "ưỡn"},
	{"r3a ượn", "ượn", "ượn"},
	// All 5 tones on "uôn"
	{"r3a uồn", "uồn", "uồn"},
	{"r3a uốn", "uốn", "uốn"},
	{"r3a uổn", "uổn", "uổn"},
	{"r3a uỗn", "uỗn", "uỗn"},
	{"r3a uộn", "uộn", "uộn"},

	// ════════════════════════════════════════════════════════════════════════
	// Rule 3b — Diphthong ia/ya/ua/ưa WITHOUT coda: tone on 1st letter
	// ════════════════════════════════════════════════════════════════════════
	{"r3b ỉa", "ỉa", "ỉa"},
	{"r3b tủa", "tủa", "tủa"},
	{"r3b cứa", "cứa", "cứa"},
	{"r3b thùa", "thùa", "thùa"},
	{"r3b khứa", "khứa", "khứa"},
	{"r3b ứa", "ứa", "ứa"},
	{"r3b ừa", "ừa", "ừa"},

	// ════════════════════════════════════════════════════════════════════════
	// Rule 4a — "ia": gi-initial → tone on a; else → tone on i
	// ════════════════════════════════════════════════════════════════════════
	{"r4a già", "già", "già"},
	{"r4a giá", "giá", "giá"},
	{"r4a giả", "giả", "giả"},
	{"r4a bịa", "bịa", "bịa"},
	{"r4a chìa", "chìa", "chìa"},
	{"r4a tía", "tía", "tía"},
	// Special case: gịa (giặt gịa) — g-initial, ia diphthong, tone on i
	{"r4a gịa special", "gịa", "gịa"},

	// ════════════════════════════════════════════════════════════════════════
	// Rule 4b — "ua": qu-initial → tone on a; else → tone on u
	// ════════════════════════════════════════════════════════════════════════
	{"r4b quán", "quán", "quán"},
	{"r4b quà", "quà", "quà"},
	{"r4b quạ", "quạ", "quạ"},
	{"r4b túa", "túa", "túa"},
	{"r4b múa", "múa", "múa"},
	{"r4b chùa", "chùa", "chùa"},

	// ════════════════════════════════════════════════════════════════════════
	// No-tone (ngang) — must be returned unchanged
	// ════════════════════════════════════════════════════════════════════════
	{"ngang ban", "ban", "ban"},
	{"ngang hoa", "hoa", "hoa"},
	{"ngang qua", "qua", "qua"},
	{"ngang gia", "gia", "gia"},
	{"ngang toan", "toan", "toan"},
	{"ngang chien", "chien", "chien"},

	// ════════════════════════════════════════════════════════════════════════
	// No-initial, no-coda — single vowel syllables
	// ════════════════════════════════════════════════════════════════════════
	{"lone ổ", "ổ", "ổ"},
	{"lone ứ", "ứ", "ứ"},
	{"lone ề", "ề", "ề"},
	// No initial, with coda
	{"lone ổn", "ổn", "ổn"},
	{"lone ẩm", "ẩm", "ẩm"},
	{"lone ấm", "ấm", "ấm"},
	// No initial, semivowel coda
	{"lone ào", "ào", "ào"},
	// No initial, uy cluster — no onset → u is nucleus, tone stays on u
	{"lone ủy", "ủy", "ủy"},
	// Contrast: uy WITH initial consonant → u is medial glide, tone on y (§VI)
	{"onset thuỷ", "thuỷ", "thuỷ"},
	{"onset nguỵ", "nguỵ", "nguỵ"},

	// ════════════════════════════════════════════════════════════════════════
	// Mixed-case and uppercase — case must be preserved
	// ════════════════════════════════════════════════════════════════════════
	{"case Hoà", "Hoà", "Hoà"},
	{"case TIẾNG", "TIẾNG", "TIẾNG"},
	{"case Việt", "Việt", "Việt"},
	{"case Thành", "Thành", "Thành"},

	// ════════════════════════════════════════════════════════════════════════
	// Multi-word sentences
	// ════════════════════════════════════════════════════════════════════════
	{"sent basic", "tiếng Việt rất đẹp", "tiếng Việt rất đẹp"},
	{"sent chuyến", "chuyến thuyền huyền bí", "chuyến thuyền huyền bí"},
	{"sent thành phố", "thành phố", "thành phố"},
	{"sent Hà Nội", "Hà Nội", "Hà Nội"},
	{"sent thủ đô", "thủ đô", "thủ đô"},
	{"sent đường phố", "đường phố", "đường phố"},
	{"sent ủy ban", "ủy ban", "ủy ban"},
	{"sent xe buýt", "xe buýt", "xe buýt"},
	{"sent quận huyện", "quận huyện", "quận huyện"},
	{"sent người Việt", "người Việt", "người Việt"},
}

func TestNormalizeToneMarks(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeToneMarks(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeToneMarks(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════════════════
// Additional test cases
// ════════════════════════════════════════════════════════════════════════════

var normalizationCases = []struct {
	name  string
	input string // tone placed at WRONG position
	want  string // tone placed at CORRECT position
}{
	// ── iê/yê diphthong + coda: tone wrongly on 1st letter → move to 2nd ────
	{"norm iê: tíên→tiến", "tíên", "tiến"},
	{"norm iê: víêng→viếng", "víêng", "viếng"},
	{"norm iê: chíêm→chiếm", "chíêm", "chiếm"},
	{"norm iê: thíên→thiến", "thíên", "thiến"},
	{"norm yê: ýêu→yếu", "ýêu", "yếu"},
	// medial + yê + coda: tone wrongly on medial u
	{"norm medial+yê: chúyên→chuyến", "chúyên", "chuyến"},
	{"norm medial+yê: thúyên→thuyến", "thúyên", "thuyến"},

	// ── ươ diphthong + coda: tone wrongly on 1st letter (ư) → move to 2nd (ơ)
	{"norm ươ: ứơn→ướn", "ứơn", "ướn"},
	{"norm ươ: cứơng→cướng", "cứơng", "cướng"},
	{"norm ươ: tứơng→tướng", "tứơng", "tướng"},
	{"norm ươ: đứơng→đướng", "đứơng", "đướng"},

	// ── uô diphthong + coda: tone wrongly on 1st letter (u) → move to 2nd (ô)
	{"norm uô: múôn→muốn", "múôn", "muốn"},
	{"norm uô: búôn→buôn (ngang stays)", "buôn", "buôn"}, // ngang — no movement
	{"norm uô: núôi→nuôi (ngang stays)", "nuôi", "nuôi"}, // ngang — no movement

	// ── oa + coda: tone wrongly on medial o → move to nucleus a ─────────────
	{"norm oa: hòan→hoàn", "hòan", "hoàn"},
	{"norm oa: ngòai→ngoài", "ngòai", "ngoài"},
	{"norm oa: lòan→loàn", "lòan", "loàn"},

	// ── ua no-coda, non-qu: tone wrongly on a → move to u ────────────────────
	{"norm ua: tuá→túa", "tuá", "túa"},
	{"norm ua: muá→múa", "muá", "múa"},
	{"norm ua: chuà→chùa (already correct)", "chùa", "chùa"},

	// ── ia no-initial: tone wrongly on a → move to i ─────────────────────────
	{"norm ia: iả→ỉa", "iả", "ỉa"},
	{"norm ia: iá→ía", "iá", "ía"},

	// ── ưa no-coda: tone wrongly on a → move to ư ────────────────────────────
	{"norm ưa: ưá→ứa", "ưá", "ứa"},
	{"norm ưa: ưà→ừa", "ưà", "ừa"},

	// ── uy with onset: tone wrongly on u → move to y ─────────────────────────
	{"norm uy onset: thủy→thuỷ", "thủy", "thuỷ"},
	{"norm uy onset: ngủy→nguỷ", "ngủy", "nguỷ"},
	{"norm uy onset: sủyt→suỷt", "sủyt", "suỷt"},
}

// ── Additional idempotency cases ─────────────────────────────────────────────

var additionalIdempotentCases = []struct {
	name  string
	input string
	want  string
}{
	// tr- initial
	{"tr trường", "trường", "trường"},
	{"tr trước", "trước", "trước"},
	{"tr trời", "trời", "trời"},
	{"tr tránh", "tránh", "tránh"},

	// ph- initial
	{"ph phở", "phở", "phở"},
	{"ph phải", "phải", "phải"},
	{"ph phức", "phức", "phức"},
	{"ph Phòng", "Phòng", "Phòng"},

	// kh- initial
	{"kh khoá", "khoá", "khoá"},
	{"kh khuyết", "khuyết", "khuyết"},

	// Huế: h + u(medial) + ê — tone on ê (medial glide present)
	{"Huế", "Huế", "Huế"},

	// Vietnamese place names
	{"place Đà Nẵng", "Đà Nẵng", "Đà Nẵng"},
	{"place Hải Phòng", "Hải Phòng", "Hải Phòng"},
	{"place Cần Thơ", "Cần Thơ", "Cần Thơ"},
	{"place Huế city", "thành phố Huế", "thành phố Huế"},
	{"place Biên Hoà", "Biên Hoà", "Biên Hoà"},

	// ia no-initial: tone correctly on i
	{"ia standalone ía", "ía", "ía"},
	{"ia standalone ĩa", "ĩa", "ĩa"},

	// ưa no-coda: tone on ư
	{"ưa ứa", "ứa", "ứa"},
	{"ưa ừa", "ừa", "ừa"},
	{"ưa ửa", "ửa", "ửa"},

	// iê no-initial + coda
	{"iê no-init iếc", "iếc", "iếc"},
	{"iê no-init iệu", "iệu", "iệu"},

	// All 5 tones on "tiêX" structure
	{"tiê tiền", "tiền", "tiền"},
	{"tiê tiến", "tiến", "tiến"},
	{"tiê tiểu", "tiểu", "tiểu"},
	{"tiê tiễn", "tiễn", "tiễn"},
	{"tiê tiện", "tiện", "tiện"},

	// All 5 tones on ươ+ng
	{"ương cường", "cường", "cường"},
	{"ương cướng", "cướng", "cướng"},
	{"ương cưởng", "cưởng", "cưởng"},
	{"ương cưỡng", "cưỡng", "cưỡng"},
	{"ương cượng", "cượng", "cượng"},

	// uy with onset, all 5 tones
	{"uy onset thuỳ", "thuỳ", "thuỳ"},
	{"uy onset thuý", "thuý", "thuý"},
	{"uy onset thuỷ", "thuỷ", "thuỷ"},
	{"uy onset thuỹ", "thuỹ", "thuỹ"},
	{"uy onset thuỵ", "thuỵ", "thuỵ"},

	// uy without onset, all 5 tones (tone must stay on u)
	{"uy no-onset ùy", "ùy", "ùy"},
	{"uy no-onset úy", "úy", "úy"},
	{"uy no-onset ủy", "ủy", "ủy"},
	{"uy no-onset ũy", "ũy", "ũy"},
	{"uy no-onset ụy", "ụy", "ụy"},

	// Long real-world sentences
	{"sent province names", "tỉnh Bình Dương", "tỉnh Bình Dương"},
	{"sent full address", "đường Nguyễn Huệ quận một", "đường Nguyễn Huệ quận một"},
	{"sent Việt Nam", "nước Việt Nam", "nước Việt Nam"},
}

func TestNormalizeToneMarksNormalization(t *testing.T) {
	for _, tc := range normalizationCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeToneMarks(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeToneMarks(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeToneMarksAdditional(t *testing.T) {
	for _, tc := range additionalIdempotentCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeToneMarks(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeToneMarks(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════════════════
// Ethnic minority place names — tone mark misplaced on consonant
// These are Central Highlands place names where the source data encodes the
// tone mark on a consonant letter instead of the vowel.
// ════════════════════════════════════════════════════════════════════════════

var ethnicPlaceNameCases = []struct {
	name  string
	input string
	want  string
}{
	// Tone on final consonant k (precomposed ḳ/ḱ) → move to preceding vowel
	{"Ngoḳ dot-below on k", "Ngoḳ", "Ngọk"},
	{"Buḱ acute on k", "Buḱ", "Búk"},

	// Tone on consonant r in cluster kr/hr/sr → move to following vowel
	{"Kŕai acute on r", "Kŕai", "Krái"},
	{"SŔo acute on R (uppercase)", "SŔo", "SRó"},
	{"Hŕu acute on r", "Hŕu", "Hrú"},

	// Tone on consonant n in cluster kn → move to following vowel
	{"Kńôp acute on n, single vowel ô", "Kńôp", "Knốp"},
	{"Kńuêc acute on n, diphthong uê+coda → tone on ê", "Kńuêc", "Knuếc"},

	// Tone already on correct vowel — must be unchanged
	{"Tụ already correct", "Tụ", "Tụ"},
	{"Réo already correct", "Réo", "Réo"},

	// Apostrophe separator with stray tone mark (M'́Drăk)
	{"M apostrophe Drăk stray acute", "M'́Drăk", "M'Drắk"},

	// Full place name phrases
	{"Ngoḳ Bay phrase", "Ngoḳ Bay", "Ngọk Bay"},
	{"Ngoḳ Tụ phrase", "Ngoḳ Tụ", "Ngọk Tụ"},
	{"Ngoḳ Réo phrase", "Ngoḳ Réo", "Ngọk Réo"},
	{"Ia Kŕai phrase", "Ia Kŕai", "Ia Krái"},
	{"SŔo phrase", "SŔo", "SRó"},
	{"Ia Hŕu phrase", "Ia Hŕu", "Ia Hrú"},
	{"Krông Buḱ phrase", "Krông Buḱ", "Krông Búk"},
	{"Ea Kńôp phrase", "Ea Kńôp", "Ea Knốp"},
	{"M'́Drăk phrase", "M'́Drăk", "M'Drắk"},
	{"Ea Kńuêc phrase", "Ea Kńuêc", "Ea Knuếc"},
}

func TestNormalizeToneMarksEthnicPlaceNames(t *testing.T) {
	for _, tc := range ethnicPlaceNameCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeToneMarks(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeToneMarks(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════════════════
// Quyết định 1989/QĐ-BGDĐT — Điều 8 official examples
// All words below are already correctly placed; NormalizeToneMarks must be
// idempotent on them (input == output).
// ════════════════════════════════════════════════════════════════════════════

var dieuEightOfficialExamples = []struct {
	name  string
	input string
	want  string
}{
	// ── §1.1 — Single-vowel nucleus ─────────────────────────────────────────
	// Source: "mái nhà, hoà nhạc, quý hoá, thuỷ thủ, mạnh khoẻ, trí tuệ,
	//          Luật Sở hữu trí tuệ"
	{"§1.1 mái", "mái", "mái"},
	{"§1.1 nhà", "nhà", "nhà"},
	{"§1.1 hoà", "hoà", "hoà"},
	{"§1.1 nhạc", "nhạc", "nhạc"},
	{"§1.1 quý", "quý", "quý"},
	{"§1.1 hoá", "hoá", "hoá"},
	{"§1.1 thuỷ", "thuỷ", "thuỷ"},
	{"§1.1 thủ", "thủ", "thủ"},
	{"§1.1 mạnh", "mạnh", "mạnh"},
	{"§1.1 khoẻ", "khoẻ", "khoẻ"},
	{"§1.1 trí", "trí", "trí"},
	{"§1.1 tuệ", "tuệ", "tuệ"},
	{"§1.1 Luật", "Luật", "Luật"},
	{"§1.1 Sở", "Sở", "Sở"},
	{"§1.1 hữu", "hữu", "hữu"},
	// Full phrases from §1.1
	{"§1.1 phrase mái nhà", "mái nhà", "mái nhà"},
	{"§1.1 phrase hoà nhạc", "hoà nhạc", "hoà nhạc"},
	{"§1.1 phrase quý hoá", "quý hoá", "quý hoá"},
	{"§1.1 phrase thuỷ thủ", "thuỷ thủ", "thuỷ thủ"},
	{"§1.1 phrase mạnh khoẻ", "mạnh khoẻ", "mạnh khoẻ"},
	{"§1.1 phrase trí tuệ", "trí tuệ", "trí tuệ"},
	{"§1.1 phrase Luật Sở hữu trí tuệ", "Luật Sở hữu trí tuệ", "Luật Sở hữu trí tuệ"},

	// ── §1.2a — Diphthong ia/ua/ưa, no coda → tone on 1st letter ───────────
	// Source: "bìa, lụa, lửa"
	{"§1.2a bìa", "bìa", "bìa"},
	{"§1.2a lụa", "lụa", "lụa"},
	{"§1.2a lửa", "lửa", "lửa"},

	// ── §1.2b — Diphthong iê/yê/uô/uơ, with coda → tone on 2nd letter ──────
	// Source: "biển, thuyền, nhuộm, được"
	{"§1.2b biển", "biển", "biển"},
	{"§1.2b thuyền", "thuyền", "thuyền"},
	{"§1.2b nhuộm", "nhuộm", "nhuộm"},
	{"§1.2b được", "được", "được"},
	// Full phrase from §1.2b
	{"§1.2b phrase biển thuyền nhuộm được", "biển thuyền nhuộm được", "biển thuyền nhuộm được"},
}

func TestNormalizeToneMarksDieuEightOfficial(t *testing.T) {
	for _, tc := range dieuEightOfficialExamples {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeToneMarks(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeToneMarks(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

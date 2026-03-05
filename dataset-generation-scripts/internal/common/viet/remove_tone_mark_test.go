package viet

import "testing"

var removeToneMarkCases = []struct {
	name  string
	input string
	want  string
}{
	// ════════════════════════════════════════════════════════════════════════
	// Empty and plain ASCII — must pass through unchanged
	// ════════════════════════════════════════════════════════════════════════
	{"empty string", "", ""},
	{"plain ascii lower", "ban", "ban"},
	{"plain ascii upper", "VIET", "VIET"},
	{"plain ascii mixed", "Hello", "Hello"},
	{"digits and punctuation", "năm 2024, tháng 12!", "nam 2024, thang 12!"},

	// ════════════════════════════════════════════════════════════════════════
	// đ / Đ — no Unicode decomposition, handled by explicit replacement
	// ════════════════════════════════════════════════════════════════════════
	{"đ → d", "đ", "d"},
	{"Đ → D", "Đ", "D"},
	{"đường", "đường", "duong"},
	{"Đà", "Đà", "Da"},

	// ════════════════════════════════════════════════════════════════════════
	// All 5 tones on plain vowels (a, e, i, o, u, y)
	// ════════════════════════════════════════════════════════════════════════
	{"à→a huyền", "à", "a"},
	{"á→a sắc", "á", "a"},
	{"ả→a hỏi", "ả", "a"},
	{"ã→a ngã", "ã", "a"},
	{"ạ→a nặng", "ạ", "a"},
	{"è→e", "è", "e"},
	{"é→e", "é", "e"},
	{"ì→i", "ì", "i"},
	{"í→i", "í", "i"},
	{"ò→o", "ò", "o"},
	{"ó→o", "ó", "o"},
	{"ù→u", "ù", "u"},
	{"ú→u", "ú", "u"},
	{"ỳ→y", "ỳ", "y"},
	{"ý→y", "ý", "y"},

	// ════════════════════════════════════════════════════════════════════════
	// Modified base vowels (no tone) — diacritic stripped
	// ════════════════════════════════════════════════════════════════════════
	{"â→a", "â", "a"},
	{"ă→a", "ă", "a"},
	{"ê→e", "ê", "e"},
	{"ô→o", "ô", "o"},
	{"ơ→o", "ơ", "o"},
	{"ư→u", "ư", "u"},

	// ════════════════════════════════════════════════════════════════════════
	// Modified vowels WITH tone marks (double diacritic → plain base)
	// ════════════════════════════════════════════════════════════════════════
	// â + all 5 tones
	{"ầ→a", "ầ", "a"},
	{"ấ→a", "ấ", "a"},
	{"ẩ→a", "ẩ", "a"},
	{"ẫ→a", "ẫ", "a"},
	{"ậ→a", "ậ", "a"},
	// ă + all 5 tones
	{"ằ→a", "ằ", "a"},
	{"ắ→a", "ắ", "a"},
	{"ẳ→a", "ẳ", "a"},
	{"ẵ→a", "ẵ", "a"},
	{"ặ→a", "ặ", "a"},
	// ê + all 5 tones
	{"ề→e", "ề", "e"},
	{"ế→e", "ế", "e"},
	{"ể→e", "ể", "e"},
	{"ễ→e", "ễ", "e"},
	{"ệ→e", "ệ", "e"},
	// ô + all 5 tones
	{"ồ→o", "ồ", "o"},
	{"ố→o", "ố", "o"},
	{"ổ→o", "ổ", "o"},
	{"ỗ→o", "ỗ", "o"},
	{"ộ→o", "ộ", "o"},
	// ơ + all 5 tones
	{"ờ→o", "ờ", "o"},
	{"ớ→o", "ớ", "o"},
	{"ở→o", "ở", "o"},
	{"ỡ→o", "ỡ", "o"},
	{"ợ→o", "ợ", "o"},
	// ư + all 5 tones
	{"ừ→u", "ừ", "u"},
	{"ứ→u", "ứ", "u"},
	{"ử→u", "ử", "u"},
	{"ữ→u", "ữ", "u"},
	{"ự→u", "ự", "u"},

	// ════════════════════════════════════════════════════════════════════════
	// Full syllables
	// ════════════════════════════════════════════════════════════════════════
	{"tiếng", "tiếng", "tieng"},
	{"Việt", "Việt", "Viet"},
	{"người", "người", "nguoi"},
	{"đường", "đường", "duong"},
	{"Nguyễn", "Nguyễn", "Nguyen"},
	{"chuyển", "chuyển", "chuyen"},
	{"thuyền", "thuyền", "thuyen"},
	{"mượn", "mượn", "muon"},
	{"phường", "phường", "phuong"},
	{"quận", "quận", "quan"},
	{"huyện", "huyện", "huyen"},

	// ════════════════════════════════════════════════════════════════════════
	// Uppercase and mixed case
	// ════════════════════════════════════════════════════════════════════════
	{"TIẾNG", "TIẾNG", "TIENG"},
	{"Thành", "Thành", "Thanh"},
	{"ĐƯỜNG", "ĐƯỜNG", "DUONG"},
	{"HÀ NỘI", "HÀ NỘI", "HA NOI"},

	// ════════════════════════════════════════════════════════════════════════
	// Vietnamese place names
	// ════════════════════════════════════════════════════════════════════════
	{"Hà Nội", "Hà Nội", "Ha Noi"},
	{"Đà Nẵng", "Đà Nẵng", "Da Nang"},
	{"Hải Phòng", "Hải Phòng", "Hai Phong"},
	{"Cần Thơ", "Cần Thơ", "Can Tho"},
	{"Huế", "Huế", "Hue"},
	{"Biên Hoà", "Biên Hoà", "Bien Hoa"},
	{"Biên Hòa", "Biên Hòa", "Bien Hoa"},
	{"Vũng Tàu", "Vũng Tàu", "Vung Tau"},
	{"Buôn Ma Thuột", "Buôn Ma Thuột", "Buon Ma Thuot"},
	{"Thành phố Hồ Chí Minh", "Thành phố Hồ Chí Minh", "Thanh pho Ho Chi Minh"},

	// ════════════════════════════════════════════════════════════════════════
	// Sentences
	// ════════════════════════════════════════════════════════════════════════
	{"nước Việt Nam", "nước Việt Nam", "nuoc Viet Nam"},
	{"ủy ban nhân dân", "ủy ban nhân dân", "uy ban nhan dan"},
	{"đường phố", "đường phố", "duong pho"},
	{"tỉnh Bình Dương", "tỉnh Bình Dương", "tinh Binh Duong"},
	{"quận Bình Thạnh", "quận Bình Thạnh", "quan Binh Thanh"},
	{"thành phố trực thuộc trung ương", "thành phố trực thuộc trung ương", "thanh pho truc thuoc trung uong"},

	// ════════════════════════════════════════════════════════════════════════
	// Latin diacritics (non-Vietnamese) — also stripped, by design
	// ════════════════════════════════════════════════════════════════════════
	{"café→cafe", "café", "cafe"},
	{"naïve→naive", "naïve", "naive"},
	{"résumé→resume", "résumé", "resume"},
}

func TestRemoveVietToneMark(t *testing.T) {
	for _, tc := range removeToneMarkCases {
		t.Run(tc.name, func(t *testing.T) {
			got := RemoveVietToneMark(tc.input)
			if got != tc.want {
				t.Errorf("RemoveVietToneMark(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

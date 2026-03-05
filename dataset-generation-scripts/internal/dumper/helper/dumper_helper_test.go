package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAdministrativeUnit_ProvinceLevel_ThanhPho(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Thành phố Hà Nội",
			input:    "Thành phố Hà Nội",
			expected: 1,
		},
		{
			name:     "Thành phố Hồ Chí Minh",
			input:    "Thành phố Hồ Chí Minh",
			expected: 1,
		},
		{
			name:     "thành phố (lowercase)",
			input:    "thành phố Cần Thơ",
			expected: 1,
		},
		{
			name:     "THÀNH PHỐ (uppercase)",
			input:    "THÀNH PHỐ HẢI PHÒNG",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdministrativeUnit_ProvinceLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdministrativeUnit_ProvinceLevel_Tinh(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Tỉnh Khánh Hòa",
			input:    "Tỉnh Khánh Hòa",
			expected: 2,
		},
		{
			name:     "tỉnh (lowercase)",
			input:    "tỉnh Bình Dương",
			expected: 2,
		},
		{
			name:     "TỈNH (uppercase)",
			input:    "TỈNH THANH HÓA",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdministrativeUnit_ProvinceLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdministrativeUnit_ProvinceLevel_UnknownPanic(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedMsg string
	}{
		{
			name:        "Quận Ba Đính",
			input:       "Quận Ba Đình",
			expectedMsg: "Unable to determine administrative unit name from province: Quận Ba Đình",
		},
		{
			name:        "Huyện Bình Thạnh",
			input:       "Huyện Bình Thạnh",
			expectedMsg: "Unable to determine administrative unit name from province: Huyện Bình Thạnh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PanicsWithValue(t, tt.expectedMsg, func() {
				GetAdministrativeUnit_ProvinceLevel(tt.input)
			})
		})
	}
}

func TestGetAdministrativeUnit_WardLevel_Phuong(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Phường Bắc Sơn",
			input:    "Phường Bắc Sơn",
			expected: 3,
		},
		{
			name:     "phường (lowercase)",
			input:    "phường Tân Bình",
			expected: 3,
		},
		{
			name:     "PHƯỜNG (uppercase)",
			input:    "PHƯỜNG PHÚ NHUẬN",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdministrativeUnit_WardLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdministrativeUnit_WardLevel_Xa(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Xã Tân Xã",
			input:    "Xã Tân Xã",
			expected: 4,
		},
		{
			name:     "xã (lowercase)",
			input:    "xã Bình Dương",
			expected: 4,
		},
		{
			name:     "XÃ (uppercase)",
			input:    "XÃ PHÚC THỌ",
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdministrativeUnit_WardLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdministrativeUnit_WardLevel_DacKhu(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Đặc khu",
			input:    "Đặc khu",
			expected: 5,
		},
		{
			name:     "đặc khu (lowercase)",
			input:    "đặc khu kinh tế",
			expected: 5,
		},
		{
			name:     "ĐẶC KHU (uppercase)",
			input:    "ĐẶC KHU VŨNG TÀU",
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAdministrativeUnit_WardLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdministrativeUnit_WardLevel_UnknownPanic(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedMsg string
	}{
		{
			name:        "Quận 1",
			input:       "Quận 1",
			expectedMsg: "Unable to determine administrative unit name from ward: Quận 1",
		},
		{
			name:        "Thị trấn",
			input:       "Thị trấn Châu Thành",
			expectedMsg: "Unable to determine administrative unit name from ward: Thị trấn Châu Thành",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PanicsWithValue(t, tt.expectedMsg, func() {
				GetAdministrativeUnit_WardLevel(tt.input)
			})
		})
	}
}

func TestToCodeName_SpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Hyphen separator",
			input:    "Phong Nha - Kẻ Bàng",
			expected: "phong_nha_ke_bang",
		},
		{
			name:     "Apostrophe",
			input:    "N'Iêng",
			expected: "nieng",
		},
		{
			name:     "Multiple apostrophes",
			input:    "Ea H'MLay",
			expected: "ea_hmlay",
		},
		{
			name:     "Mixed special chars",
			input:    "H'Leo - Quận",
			expected: "hleo_quan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCodeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToCodeName_ToneRemoval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Thanh Hoá",
			input:    "Thanh Hoá",
			expected: "thanh_hoa",
		},
		{
			name:     "Khánh Hòa",
			input:    "Khánh Hòa",
			expected: "khanh_hoa",
		},
		{
			name:     "Bắc Sơn",
			input:    "Bắc Sơn",
			expected: "bac_son",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCodeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToCodeName_Lowercase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Uppercase input",
			input:    "HÀ NỘI",
			expected: "ha_noi",
		},
		{
			name:     "Mixed case input",
			input:    "Hà Nội",
			expected: "ha_noi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCodeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollapseSpaces_Multiple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Two spaces become one space",
			input:    "Ninh  Binh",
			expected: "Ninh Binh",
		},
		{
			name:     "Three spaces become one (using strings.Fields)",
			input:    "Quang   Ninh",
			expected: "Quang Ninh",
		},
		{
			name:     "Multiple double spaces all collapsed",
			input:    "Ha  Noi  -  Viet  Nam",
			expected: "Ha Noi - Viet Nam",
		},
		{
			name:     "Many spaces become one",
			input:    "A    B     C",
			expected: "A B C",
		},
		{
			name:     "Tabs and spaces",
			input:    "Hello\t\tWorld",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CollapseSpaces(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollapseSpaces_Trailing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Leading space",
			input:    " Ninh Binh",
			expected: "Ninh Binh",
		},
		{
			name:     "Trailing space",
			input:    "Quang Ninh ",
			expected: "Quang Ninh",
		},
		{
			name:     "Both leading and trailing",
			input:    " Vinh Phuc ",
			expected: "Vinh Phuc",
		},
		{
			name:     "Newlines and tabs",
			input:    "\n\t  Hello World  \t\n",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CollapseSpaces(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
package dumper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/helper"
)


func TestGetAdministrativeUnit_ProvinceLevel(t *testing.T) {
	assert.Equal(t, 1, helper.GetAdministrativeUnit_ProvinceLevel("Thành phố Hồ Chí Minh"))
	assert.Equal(t, 2, helper.GetAdministrativeUnit_ProvinceLevel("Tỉnh Khánh Hoà"))
	assert.Equal(t, 2, helper.GetAdministrativeUnit_ProvinceLevel("Tỉnh Thành phố"))
}

func TestGetAdministrativeUnit_ProvinceLevel_ShouldThrowException(t *testing.T) {
	assert.PanicsWithValue(
		t, "Unable to determine administrative unit name from province: Quận Ba Đình",
		func() {
			helper.GetAdministrativeUnit_ProvinceLevel("Quận Ba Đình")
	})
}

func TestGetAdministrativeUnit_WardLevel(t *testing.T) {
	assert.Equal(t, 8, helper.GetAdministrativeUnit_WardLevel("Phường Phường Đúc"))
	assert.Equal(t, 9, helper.GetAdministrativeUnit_WardLevel("Thị trấn Châu Thành"))
	assert.Equal(t, 10, helper.GetAdministrativeUnit_WardLevel("Xã Tân Xã"))
}

func TestGetAdministrativeUnit_WardLevel_Exception(t *testing.T) {
	assert.PanicsWithValue(
		t, "Unable to determine administrative unit name from ward: Quận 9",
		func() {
			helper.GetAdministrativeUnit_WardLevel("Quận 9")
	})
}

// Test string normalization

func TestNormalizeString(t *testing.T) {
	assert.Equal(t, "Da Lat", helper.NormalizeString("Đà Lạt"))
	assert.Equal(t, "Hoi An", helper.NormalizeString("Hội An"))
	assert.Equal(t, "Bai bien Cua Lo", helper.NormalizeString("Bãi biển Cửa Lò"))
	assert.Equal(t, "Ghenh da Sen Thang Bang", helper.NormalizeString("Ghềnh đá Sên Thắng Bâng"))
	assert.Equal(t, "Ong Ot Enh Uong", helper.NormalizeString("Ông Ớt Ễnh Ương"))
}

func TestToCodeName(t *testing.T) {
	assert.Equal(t, "phong_nha_ke_bang", helper.ToCodeName("Phong Nha - Kẻ Bàng"))
	assert.Equal(t, "nieng", helper.ToCodeName("N'Iêng"))
	assert.Equal(t, "thanh_hoa", helper.ToCodeName("Thanh Hoá"))
}

func TestRemoveWhiteSpaces(t *testing.T) {
	assert.Equal(t, "Ninh Binh", helper.RemoveWhiteSpaces(" Ninh  Binh "))
	assert.Equal(t, "Quang Ninh", helper.RemoveWhiteSpaces(" Quang Ninh"))
	assert.Equal(t, "Vinh Phuc", helper.RemoveWhiteSpaces("Vinh Phuc "))
}

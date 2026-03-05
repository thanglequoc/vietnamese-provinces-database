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
	assert.Equal(t, 3, helper.GetAdministrativeUnit_WardLevel("Phường Phường Đúc"))
	assert.Equal(t, 4, helper.GetAdministrativeUnit_WardLevel("Xã Tân Xã"))
	assert.Equal(t, 5, helper.GetAdministrativeUnit_WardLevel("Đặc khu"))
}

func TestGetAdministrativeUnit_WardLevel_Exception(t *testing.T) {
	assert.PanicsWithValue(
		t, "Unable to determine administrative unit name from ward: Quận 9",
		func() {
			helper.GetAdministrativeUnit_WardLevel("Quận 9")
		})
}

// Note: normalizeString is a private function in the service package
// This test is removed as it tests a non-public API
// The function is tested indirectly through integration tests

func TestToCodeName(t *testing.T) {
	assert.Equal(t, "phong_nha_ke_bang", helper.ToCodeName("Phong Nha - Kẻ Bàng"))
	assert.Equal(t, "nieng", helper.ToCodeName("N'Iêng"))
	assert.Equal(t, "thanh_hoa", helper.ToCodeName("Thanh Hoá"))
}

func TestCollapseSpaces(t *testing.T) {
	assert.Equal(t, "Ninh Binh", helper.CollapseSpaces(" Ninh  Binh "))
	assert.Equal(t, "Quang Ninh", helper.CollapseSpaces(" Quang Ninh"))
	assert.Equal(t, "Vinh Phuc", helper.CollapseSpaces("Vinh Phuc "))
}
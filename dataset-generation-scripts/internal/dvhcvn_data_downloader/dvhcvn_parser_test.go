package dvhcvn_data_downloader

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToDvhcvnWardModel(t *testing.T) {

	input := `<TABLE diffgr:id="TABLE4405" msdata:rowOrder="4404">
	<MaTinh>36</MaTinh>
	<TenTinh>Tỉnh Nam Định</TenTinh>
	<MaQuanHuyen>360</MaQuanHuyen>
	<TenQuanHuyen>Huyện Ý Yên</TenQuanHuyen>
	<MaPhuongXa>13870</MaPhuongXa>
	<TenPhuongXa>Xã Yên Cường</TenPhuongXa>
	<LoaiHinh>Xã</LoaiHinh>
	</TABLE>`

	actualModel := toDvhcvnWardModel(input)

	assert.Equal(t, "36", actualModel.ProvinceCode)
	assert.Equal(t, "Tỉnh Nam Định", actualModel.ProvinceName)
	assert.Equal(t, "360", actualModel.DistrictCode)
	assert.Equal(t, "Huyện Ý Yên", actualModel.DistrictName)
	assert.Equal(t, "13870", actualModel.WardCode)
	assert.Equal(t, "Xã Yên Cường", actualModel.WardName)
	assert.Equal(t, "Xã", actualModel.Unit)

}

func TestExtractWardDvhcvnUnits(t *testing.T) {
	input := `<TABLE diffgr:id="TABLE6037" msdata:rowOrder="6036">
		<MaTinh>44</MaTinh>
		<TenTinh>Tỉnh Quảng Bình</TenTinh>
		<MaQuanHuyen>456</MaQuanHuyen>
		<TenQuanHuyen>Huyện Quảng Ninh</TenQuanHuyen>
		<MaPhuongXa>19240</MaPhuongXa>
		<TenPhuongXa>Xã An Ninh</TenPhuongXa>
		<LoaiHinh>Xã</LoaiHinh>
	</TABLE>
	<TABLE diffgr:id="TABLE6038" msdata:rowOrder="6037">
		<MaTinh>44</MaTinh>
		<TenTinh>Tỉnh Quảng Bình</TenTinh>
		<MaQuanHuyen>456</MaQuanHuyen>
		<TenQuanHuyen>Huyện Quảng Ninh</TenQuanHuyen>
		<MaPhuongXa>19243</MaPhuongXa>
		<TenPhuongXa>Xã Vạn Ninh</TenPhuongXa>
		<LoaiHinh>Xã</LoaiHinh>
	</TABLE>
	<TABLE diffgr:id="TABLE6039" msdata:rowOrder="6038">
		<MaTinh>44</MaTinh>
		<TenTinh>Tỉnh Quảng Bình</TenTinh>
		<MaQuanHuyen>457</MaQuanHuyen>
		<TenQuanHuyen>Huyện Lệ Thủy</TenQuanHuyen>
		<MaPhuongXa>19246</MaPhuongXa>
		<TenPhuongXa>Thị trấn NT Lệ Ninh</TenPhuongXa>
		<LoaiHinh>Thị trấn</LoaiHinh>
	</TABLE>`

	actualModels := extractWardDvhcvnUnits(input)
	assert.Equal(t, 3, len(actualModels))
	assert.Equal(t, "Xã An Ninh", actualModels[0].WardName)
	assert.Equal(t, "Xã Vạn Ninh", actualModels[1].WardName)
	assert.Equal(t, "Thị trấn NT Lệ Ninh", actualModels[2].WardName)
}

func TestConvertStandardUnitName(t *testing.T) {
	assert.Equal(t, "Thành phố Hải Dương", convertStandardUnitName("Thành Phố Hải Dương"))
	assert.Equal(t, "Thị xã Nầm Đốm", convertStandardUnitName("Thị Xã Nầm Đốm"))
	assert.Equal(t, "Thị trấn Yên Nghĩa", convertStandardUnitName("Thị Trấn Yên Nghĩa"))
}

package dvhcvn_data_downloader

/*
Response model from the DVHCVN Soap API
Sample model response:

```

<TABLE diffgr:id="TABLE173" msdata:rowOrder="172">

	<MaTinh>01</MaTinh>
	<TenTinh>Thành phố Hà Nội</TenTinh>
	<MaQuanHuyen>017</MaQuanHuyen>
	<TenQuanHuyen>Huyện Đông Anh</TenQuanHuyen>
	<MaPhuongXa>00514</MaPhuongXa>
	<TenPhuongXa>Xã Võng La</TenPhuongXa>
	<LoaiHinh>Xã</LoaiHinh>

</TABLE>

```
*/
type DvhcvnModel struct {
	ProvinceCode string // <MaTinh>
	ProvinceName string // <TenTinh>
	DistrictCode string // <MaQuanHuyen>
	DistrictName string // <TenQuanHuyen>
	WardCode     string // <MaPhuongXa>
	WardName     string // <TenPhuongXa>
	WardUnit     string // <LoaiHinh>
}

/*
Response model from the DVHCVN Soap API: DanhMucTinh
```
<TABLE diffgr:id="TABLE2" msdata:rowOrder="1">

	<MaTinh>02</MaTinh>
	<TenTinh>Tỉnh Hà Giang</TenTinh>
	<LoaiHinh>Tỉnh</LoaiHinh>

</TABLE>
````
*/
type DvhcvnProvinceModel struct {
	ProvinceCode string // <MaTinh>
	ProvinceName string // <TenTinh>
	Unit         string // <LoaiHinh>
}

/*
Response model from the DVHCVN Soap API: DanhMucQuanHuyen
```
<TABLE diffgr:id="TABLE2" msdata:rowOrder="1">

	<MaTinh>01</MaTinh>
	<TenTinh>Thành phố Hà Nội</TenTinh>
	<MaQuanHuyen>002</MaQuanHuyen>
	<TenQuanHuyen>Quận Hoàn Kiếm</TenQuanHuyen>
	<LoaiHinh>Quận</LoaiHinh>

</TABLE>
```
*/
type DvhcvnDistrictModel struct {
	ProvinceCode string // <MaTinh>
	ProvinceName string // <TenTinh>
	DistrictCode string // <MaQuanHuyen>
	DistrictName string // <TenQuanHuyen>
	Unit         string // <LoaiHinh>
}

/*
Response model from the DVHCVN Soap API: DanhMucPhuongXa
```
<TABLE diffgr:id="TABLE7428" msdata:rowOrder="7427">

	<MaTinh>64</MaTinh>
	<TenTinh>Tỉnh Gia Lai</TenTinh>
	<MaQuanHuyen>622</MaQuanHuyen>
	<TenQuanHuyen>Thành phố Pleiku</TenQuanHuyen>
	<MaPhuongXa>23566</MaPhuongXa>
	<TenPhuongXa>Phường Hội Thương</TenPhuongXa>
	<LoaiHinh>Phường</LoaiHinh>

</TABLE>
```
*/
type DvhcvnWardModel struct {
	ProvinceCode string // <MaTinh>
	ProvinceName string // <TenTinh>
	DistrictCode string // <MaQuanHuyen>
	DistrictName string // <TenQuanHuyen>
	WardCode     string // <MaPhuongXa>
	WardName     string // <TenPhuongXa>
	Unit         string // <LoaiHinh>
}

type DvhcvnDataSet struct {
	ProvinceData []DvhcvnProvinceModel
	DistrictData []DvhcvnDistrictModel
	WardData     []DvhcvnWardModel
}

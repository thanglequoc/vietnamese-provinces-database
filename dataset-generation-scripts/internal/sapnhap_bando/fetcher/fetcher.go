package fetcher

import (
	"encoding/json"
	"net/http"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
)

const GET_ALL_PROVINCES_URL = "https://sapnhap.bando.com.vn/pcotinh"

/*
Get all the provinces data from the sapnhap site
POST: https://sapnhap.bando.com.vn/pcotinh
*/
func GetAllProvincesDataFromSapNhapSite() []dto.SapNhapProvinceData {
	res, err := http.Post(GET_ALL_PROVINCES_URL, "application/json", nil)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	var provincesData []dto.SapNhapProvinceData
	if err := json.NewDecoder(res.Body).Decode(&provincesData); err != nil {
		panic(err)
	}
	return provincesData
}

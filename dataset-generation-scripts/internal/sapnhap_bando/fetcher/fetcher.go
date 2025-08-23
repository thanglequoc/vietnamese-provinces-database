package fetcher

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
)

const GET_ALL_PROVINCES_URL = "https://sapnhap.bando.com.vn/pcotinh"
const GET_ALL_WARDS_OF_PROVINCES_URL = "https://sapnhap.bando.com.vn/ptracuu"

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

/*
API Look up to get all wards of a province from the sapnhap site
POST: https://sapnhap.bando.com.vn/ptracuu
*/
func GetAllWardsOfProvinceFromSapNhapSite(provinceID int) []dto.SapNhapWardData {
	form := url.Values{}
	form.Add("id", strconv.Itoa(provinceID))

	res, err := http.Post(GET_ALL_WARDS_OF_PROVINCES_URL, "application/x-www-form-urlencoded", bytes.NewBufferString(form.Encode()))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	var wardsData []dto.SapNhapWardData
	if err := json.NewDecoder(res.Body).Decode(&wardsData); err != nil {
		panic(err)
	}
	return wardsData
}

/*
Load the bando gis province mapping from file
*/
func LoadBanDoGISProvincesFromFile(path string) ([]dto.BanDoGISProvince, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode JSON
	var provinces []dto.BanDoGISProvince
	if err := json.Unmarshal(data, &provinces); err != nil {
		return nil, err
	}
	return provinces, nil
}

/*
Load the bando gis ward mapping from file
*/
func LoadBanDoGISWardsFromFile(path string) ([]dto.BanDoGISWard, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode JSON
	var wards []dto.BanDoGISWard
	if err := json.Unmarshal(data, &wards); err != nil {
		return nil, err
	}
	return wards, nil
}

// TODO: Add API to get JSON GIS data response

package fetcher

import "testing"

func TestGetAllProvincesDataFromSapNhapSite(t *testing.T) {
	provincesData := GetAllProvincesDataFromSapNhapSite()
	if len(provincesData) == 0 {
		panic("No provinces data fetched from sapnhap site")
	}
}

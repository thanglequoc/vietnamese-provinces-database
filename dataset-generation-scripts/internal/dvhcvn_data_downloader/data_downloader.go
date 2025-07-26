package dvhcvn_data_downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const DVHCVN_URL = "https://danhmuchanhchinh.gso.gov.vn/DMDVHC.asmx"

const DVHCVN_SOAP_DANH_MUC_TINH = "DanhMucTinh"
const DVHCVN_SOAP_DANH_MUC_QUAN_HUYEN = "DanhMucQuanHuyen"
const DVHCVN_SOAP_DANH_MUC_PHUONG_XA = "DanhMucPhuongXa"

/*
Fetch the data from the public government API url
Required the selected data date
*/
func FetchDvhcvnData(publicDataDate time.Time) DvhcvnDataSet {
	fmt.Printf("⬇️ Downloading provinces data patch from %s\n", DVHCVN_URL)
	dataAPIDateStr := publicDataDate.Format("02/01/2006") // dd/MM/YYYY

	// Fetch province-level data from DVHCVN SOAP API
	provinceRequestBody := createSoapRequest(DVHCVN_SOAP_DANH_MUC_TINH, dataAPIDateStr)
	provinceSoapResponse := fetchDataFromAPI(provinceRequestBody)
	provinceResult := extractProvinceDvhcvnUnits(provinceSoapResponse)

	// Fetch district-level data from DVHCVN SOAP API
	districtRequestBody := createSoapRequest(DVHCVN_SOAP_DANH_MUC_QUAN_HUYEN, dataAPIDateStr)
	districtSoapResponse := fetchDataFromAPI(districtRequestBody)
	districtResult := extractDistrictDvhcvnUnits(districtSoapResponse)

	// Fetch ward-level data from DVHCVN SOAP API
	wardRequestBody := createSoapRequest(DVHCVN_SOAP_DANH_MUC_PHUONG_XA, dataAPIDateStr)
	wardSoapResponse := fetchDataFromAPI(wardRequestBody)
	wardResult := extractWardDvhcvnUnits(wardSoapResponse)

	// Concatenate and return results
	return DvhcvnDataSet{
		ProvinceData: provinceResult,
		DistrictData: districtResult,
		WardData: wardResult,
	}
}

// Helper function to create a SOAP request body
func createSoapRequest(action, date string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
		<soap:Body>
			<%s xmlns="http://tempuri.org/">
				<DenNgay>%s</DenNgay>
			</%s>
		</soap:Body>
	</soap:Envelope>`, action, date, action)
}

// Helper function to send HTTP POST requests and fetch data
func fetchDataFromAPI(requestBody string) string {
	res, err := http.Post(DVHCVN_URL, "text/xml", strings.NewReader(requestBody))
	if err != nil {
		log.Fatalf("Exception occurred while making the request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Exception occurred while reading the response: %v", err)
	}

	return string(body)
}

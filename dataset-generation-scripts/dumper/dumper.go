package dumper

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	vn_common "github.com/thanglequoc-vn-provinces/v2/common"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	data_downloader "github.com/thanglequoc-vn-provinces/v2/dvhcvn_data_downloader"

	"io/fs"
	"path/filepath"
)

// Temporary deprecated, API upstream data is not up to date
func BeginDumpingDataWithDvhcvnDirectSource() {
	fmt.Print("(Optional) Please specify the data date (dd/MM/YYYY). Leave empty to go with default option: ")

	reader := bufio.NewReader(os.Stdin)

	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(userInput)
	fmt.Println("Selected date: ", userInput)

	var dataSetTime time.Time
	if len(strings.TrimSpace(userInput)) == 0 {
		fmt.Println("No input is recorded, using tomorrow as the default day...")
		dataSetTime = time.Now().Add(time.Hour * 24)
	} else {
		dataSetTime, _ = time.Parse("02/01/2006", userInput) // dd/MM/yyyy
	}

	dvhcvnUnits := data_downloader.FetchDvhcvnData(dataSetTime)

	fmt.Println(dvhcvnUnits)
	// insertToProvinces(dvhcvnUnits.ProvinceData)
	// insertToWards(dvhcvnUnits.WardData)
	fmt.Println("📥 Dumper operation finished")
}

// Dump the SQL script from the manual database degree seed
func DumpFromManualSeed() {
	fmt.Println("Dumping data from manual database degree seed...")
	vn_common.ExecuteSQLScript("./resources/manual_decree_seeds/provinces_seed.sql")

	wardSeedRootFolder := "./resources/manual_decree_seeds/wards"
	err := filepath.WalkDir(wardSeedRootFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".sql" {
			fmt.Printf("Executing script: %s\n", path)
			vn_common.ExecuteSQLScript(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking through manual decree seeds: %v\n", err)
	}

	// Thing to do: Manual inject data to the database - Done
	// Read from the insert data, construct the dvhcvn data
	seedProvinces := vn_common.GetAllSeedProvinces()
	seedWards := vn_common.GetAllSeedWards()
	insertToProvinces(seedProvinces)
	insertToWards(seedWards)
	fmt.Println("📥 Dumper operation finished")
}

func insertToWards(seedWardModels []vn_common.SeedWard) {
	db := vn_common.GetPostgresDBConnection()
	ctx := context.Background()
	totalWard := 0

	for _, w := range seedWardModels {
		wardFullName := removeWhiteSpaces(w.Name)
		administrativeUnitLevel := getAdministrativeUnit_WardLevel(wardFullName)
		unitName := AdministrativeUnitNamesShortNameMap_vn[administrativeUnitLevel]
		unitName_en := AdministrativeUnitNamesShortNameMap_en[administrativeUnitLevel]
		wardShortName := strings.Trim(strings.Replace(wardFullName, unitName, "", 1), " ")
		codeName := toCodeName(wardShortName)
		wardShortNameEn := normalizeString(wardShortName)

		// Case when ward name is a number
		isNumber, _ := regexp.MatchString("[0-9]+", wardShortName)
		var wardFullNameEn string
		if isNumber {
			wardFullNameEn = unitName_en + " " + wardShortNameEn
		} else {
			wardFullNameEn = wardShortNameEn + " " + unitName_en
		}

		wardModel := &vn_common.Ward{
			Code:                 w.Code,
			Name:                 wardShortName,
			NameEn:               wardShortNameEn,
			FullName:             wardFullName,
			FullNameEn:           wardFullNameEn,
			CodeName:             codeName,
			AdministrativeUnitId: administrativeUnitLevel,
			ProvinceCode:         w.ProvinceCode,
		}

		_, err := db.NewInsert().Model(wardModel).Exec(ctx)
		totalWard++
		if err != nil {
			fmt.Println(err)
			panic("Exception happens while inserting into wards table")
		}
	}

	fmt.Printf("Inserted %d wards to tables\n", totalWard)
}

func insertToProvinces(seedProvinceModels []vn_common.SeedProvince) {
	db := vn_common.GetPostgresDBConnection()
	ctx := context.Background()

	for _, p := range seedProvinceModels {
		provinceFullName := removeWhiteSpaces(p.Name)
		administrativeUnitLevel := getAdministrativeUnit_ProvinceLevel(provinceFullName)
		unitName := AdministrativeUnitNamesShortNameMap_vn[administrativeUnitLevel]
		unitName_en := AdministrativeUnitNamesShortNameMap_en[administrativeUnitLevel]
		provinceShortName := strings.Trim(strings.Replace(provinceFullName, unitName, "", 1), " ")
		codeName := toCodeName(provinceShortName)
		provinceShortNameEn := normalizeString(provinceShortName)
		provinceFullNameEn := provinceShortNameEn + " " + unitName_en

		provinceModel := &vn_common.Province{
			Code:                   p.Code,
			Name:                   provinceShortName,
			NameEn:                 provinceShortNameEn,
			FullName:               provinceFullName,
			FullNameEn:             provinceFullNameEn,
			CodeName:               codeName,
			AdministrativeUnitId:   administrativeUnitLevel,
		}

		_, err := db.NewInsert().Model(provinceModel).Exec(ctx)
		if err != nil {
			fmt.Println(err)
			panic("Exception happens while inserting into provinces table")
		}
	}

	fmt.Printf("Inserted %d provinces to tables\n", len(seedProvinceModels))
}

/*
Determine the province administrative unit id from its name
*/
func getAdministrativeUnit_ProvinceLevel(provinceFullName string) int {
	if strings.HasPrefix(provinceFullName, "Thành phố") {
		return 1
	}
	if strings.HasPrefix(provinceFullName, "Tỉnh") {
		return 2
	}
	panic("Unable to determine administrative unit name from province: " + provinceFullName)
}

/*
Determine the ward administrative unit id from its name
*/
func getAdministrativeUnit_WardLevel(wardFullName string) int {
	if strings.HasPrefix(wardFullName, "Phường") {
		return 3
	}
	if strings.HasPrefix(wardFullName, "Xã") {
		return 4
	}
	if strings.HasPrefix(wardFullName, "Đặc khu") {
		return 5
	}
	panic("Unable to determine administrative unit name from ward: " + wardFullName)
}

/*
Normalize string to remove Vietnamese special character and sign
*/
func normalizeString(source string) string {
	trans := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(trans, source)
	result = strings.ReplaceAll(result, "đ", "d")
	result = strings.ReplaceAll(result, "Đ", "D")
	return result
}

/*
Generate code name from the name
*/
func toCodeName(shortName string) string {
	shortName = strings.ReplaceAll(shortName, " - ", " ")
	shortName = strings.ReplaceAll(shortName, "'", "") // to handle special name with single quote
	return strings.ToLower(strings.ReplaceAll(normalizeString(shortName), " ", "_"))
}

func removeWhiteSpaces(name string) string {
	return strings.Trim(strings.ReplaceAll(name, "  ", " "), " ")
}

package dataset_writer

import (
	"bufio"
	"fmt"
	"os"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

const hsetAdministrativeUnitTemplate string = "HSET administrativeUnit:%d id %d fullName \"%s\" fullNameEn \"%s\" shortName \"%s\" shortNameEn \"%s\" codeName \"%s\"\n"
const hsetRegionTemplate string = "HSET region:%d name \"%s\" nameEn \"%s\" codeName \"%s\" \n"

const hsetProvinceTemplate string = "HSET province:%s code \"%s\" name \"%s\" nameEn \"%s\" fullName \"%s\" fullNameEn \"%s\" codeName \"%s\" postalCodePrefix \"%s\" administrativeUnitId %d \n"

const hsetWardTemplate string = "HSET ward:%s code \"%s\" name \"%s\" nameEn \"%s\" fullName \"%s\" fullNameEn \"%s\" codeName \"%s\" postalCode \"%s\" administrativeUnitId %d districtCode \"%s\" \n"
const saddProvinceWardTemplate string = "SADD province:%s:wards \"%s\" \n"
const hsetProvinceWardVnTemplate string = "HSET province:%s:wards:vn \"%s\" \"%s\" \n"
const hsetProvinceWardEnTemplate string = "HSET province:%s:wards:en \"%s\" \"%s\" \n"

type RedisDatasetFileWriter struct {
	OutputFolderPath string
}

func (w *RedisDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	os.MkdirAll(w.OutputFolderPath, 0746)

	redisDatasetFilePath := fmt.Sprintf("%s/redis_vn_provinces_dataset.redis", w.OutputFolderPath)
	redisDatasetFile, err := os.OpenFile(redisDatasetFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	dataWriter := bufio.NewWriter(redisDatasetFile)

	for _, a := range administrativeUnits {
		dataWriter.WriteString(generateAdministrativeRecord(a))
	}

	for _, r := range regions {
		dataWriter.WriteString(generateRegionRecord(r))
	}

	for _, p := range provinces {
		dataWriter.WriteString(generateProvinceRecord(p))
	}
	for _, w := range wards {
		dataWriter.WriteString(generateWardRecord(w))
		dataWriter.WriteString(generateProvinceWardRelationship(w))
	}

	dataWriter.Flush()
	redisDatasetFile.Close()

	return writeRedisReadme(w.OutputFolderPath)
}

func writeRedisReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"Redis Dataset — Vietnamese Provinces Database",
		"Redis commands loading all Vietnamese provinces, wards, regions, and administrative units.",
		[]DatasetReadmeFile{
			{Name: "redis_vn_provinces_dataset.redis", Description: "Redis HSET/SADD commands"},
		},
		[]string{
			"## Data Structure",
			"",
			"- `province:<code>` — province hash (name, nameEn, fullName, codeName, postalCodePrefix, administrativeUnitId)",
			"- `ward:<code>` — ward hash (name, fullName, codeName, postalCode, administrativeUnitId, districtCode)",
			"- `administrativeUnit:<id>` — unit type hash",
			"- `region:<id>` — region hash",
			"- `province:<code>:wards` — SET of ward codes",
			"- `province:<code>:wards:vn` / `province:<code>:wards:en` — ward code → name hashes",
			"",
			"## Sample Queries",
			"",
			"```bash",
			"redis-cli HGETALL province:01",
			"redis-cli SMEMBERS province:01:wards",
			"redis-cli HGET ward:00004 fullName",
			"```",
		})
}

func generateAdministrativeRecord(a model.AdministrativeUnit) string {
	return fmt.Sprintf(hsetAdministrativeUnitTemplate, a.Id, a.Id, a.FullName, a.FullNameEn, a.ShortName, a.ShortNameEn, a.CodeName)
}

func generateRegionRecord(r model.AdministrativeRegion) string {
	return fmt.Sprintf(hsetRegionTemplate, r.Id, r.Name, r.NameEn, r.CodeName)
}

func generateProvinceRecord(p model.Province) string {
	return fmt.Sprintf(hsetProvinceTemplate, p.Code, p.Code, p.Name, p.NameEn, p.FullName, p.FullNameEn, p.CodeName, p.PostalCodePrefix, p.AdministrativeUnitId)
}

func generateWardRecord(w model.Ward) string {
	return fmt.Sprintf(hsetWardTemplate, w.Code, w.Code, w.Name, w.NameEn, w.FullName, w.FullNameEn, w.CodeName, w.PostalCode, w.AdministrativeUnitId, w.ProvinceCode)
}

func generateProvinceWardRelationship(w model.Ward) string {
	return fmt.Sprintf(saddProvinceWardTemplate, w.ProvinceCode, w.Code) + fmt.Sprintf(hsetProvinceWardVnTemplate, w.ProvinceCode, w.Code, w.FullName) + fmt.Sprintf(hsetProvinceWardEnTemplate, w.ProvinceCode, w.Code, w.FullNameEn)
}

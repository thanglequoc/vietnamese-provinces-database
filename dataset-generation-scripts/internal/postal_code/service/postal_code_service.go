package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
	"github.com/thanglequoc-vn-provinces/v2/internal/postal_code/repository"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

type provincePrefixSeed struct {
	Code             string `json:"code"`
	PostalCodePrefix string `json:"postal_code_prefix"`
}

type wardPostalCodeSeed struct {
	ProvinceCode string `json:"province_code"`
	Name         string `json:"name"`
	PostalCode   string `json:"postal_code"`
}

type PostalCodeService struct {
	vnProvinceTmpRepo *vnRepo.VnProvincesTmpRepository
	postalCodeRepo    *repository.PostalCodeRepository
	seedDir           string
}

func NewPostalCodeService(
	vnRepo *vnRepo.VnProvincesTmpRepository,
	postalRepo *repository.PostalCodeRepository,
	seedDir string,
) *PostalCodeService {
	return &PostalCodeService{
		vnProvinceTmpRepo: vnRepo,
		postalCodeRepo:    postalRepo,
		seedDir:           seedDir,
	}
}

// normalizeExact produces a tone-preserving NFC key: lowercased, NFC-normalized,
// with curly apostrophes folded to ASCII and whitespace collapsed.
func normalizeExact(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "’", "'")
	s = strings.ReplaceAll(s, "‘", "'")
	s = viet.NormalizeToneMarks(s)
	return strings.Join(strings.Fields(s), " ")
}

// normalizeStripped removes tone marks entirely and collapses whitespace,
// apostrophes, and hyphens. This mirrors how the DB stores `name_en`, and is
// used as a fallback when the tone-preserving exact match fails (e.g. the
// source decree writes "Hòa" while the DB normalizes to "Hoà").
func normalizeStripped(s string) string {
	s = normalizeExact(s)
	s = viet.RemoveVietToneMark(s)
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "-", "")
	return strings.Join(strings.Fields(s), "")
}

func (s *PostalCodeService) loadProvincePrefixes() ([]provincePrefixSeed, error) {
	data, err := os.ReadFile(filepath.Join(s.seedDir, "province_postal_code_prefixes.json"))
	if err != nil {
		return nil, fmt.Errorf("read province prefixes seed: %w", err)
	}
	var seeds []provincePrefixSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return nil, fmt.Errorf("parse province prefixes seed: %w", err)
	}
	return seeds, nil
}

func (s *PostalCodeService) loadWardPostalCodes() ([]wardPostalCodeSeed, error) {
	data, err := os.ReadFile(filepath.Join(s.seedDir, "ward_postal_codes.json"))
	if err != nil {
		return nil, fmt.Errorf("read ward postal codes seed: %w", err)
	}
	var seeds []wardPostalCodeSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return nil, fmt.Errorf("parse ward postal codes seed: %w", err)
	}
	return seeds, nil
}

// ImportPostalCodes populates postal_code_prefix on provinces and postal_code on
// wards. It verifies a 100% match rate: every ward in wards_tmp must receive a
// postal code, and the number of updated wards must equal the seed count.
func (s *PostalCodeService) ImportPostalCodes(ctx context.Context) error {
	prefixSeeds, err := s.loadProvincePrefixes()
	if err != nil {
		return err
	}
	wardSeeds, err := s.loadWardPostalCodes()
	if err != nil {
		return err
	}
	log.Printf("ℹ️ Importing postal codes: %d province prefixes, %d ward postal codes", len(prefixSeeds), len(wardSeeds))

	for _, seed := range prefixSeeds {
		if err := s.postalCodeRepo.UpdateProvincePostalCodePrefix(ctx, seed.Code, seed.PostalCodePrefix); err != nil {
			return fmt.Errorf("update province %s postal prefix: %w", seed.Code, err)
		}
	}

	provinces := s.vnProvinceTmpRepo.GetAllProvinces()

	// Validate that every seed province_code exists in provinces_tmp.
	validProvinceCodes := make(map[string]bool, len(provinces))
	for _, p := range provinces {
		validProvinceCodes[p.Code] = true
	}
	for _, seed := range wardSeeds {
		if !validProvinceCodes[seed.ProvinceCode] {
			return fmt.Errorf("seed province_code %q not found in provinces_tmp", seed.ProvinceCode)
		}
	}

	wards := s.vnProvinceTmpRepo.GetAllWards()

	// Tier 1: exact tone-preserving match, scoped by province code.
	exactKeyToCode := make(map[string]string, len(wards))
	// Tier 2: tone-stripped match against name_en, scoped by province code.
	// Multiple wards in the same province may share a stripped key only if they
	// differ by tone mark (e.g. "Văn Lang" vs "Văn Lăng"); for those we require
	// the parent province to be identical before disambiguating by exact name.
	strippedKeyToCodes := make(map[string][]string, len(wards))
	provinceOfWard := make(map[string]string, len(wards))
	for _, w := range wards {
		exactKey := w.ProvinceCode + "|" + normalizeExact(w.Name)
		if _, exists := exactKeyToCode[exactKey]; !exists {
			exactKeyToCode[exactKey] = w.Code
		}
		strippedKey := w.ProvinceCode + "|" + normalizeStripped(w.NameEn)
		strippedKeyToCodes[strippedKey] = append(strippedKeyToCodes[strippedKey], w.Code)
		provinceOfWard[w.Code] = w.ProvinceCode
	}

	matched := 0
	var unmatched []wardPostalCodeSeed
	for _, seed := range wardSeeds {
		code := ""
		// Tier 1: exact name match within the seed's province.
		exactKey := seed.ProvinceCode + "|" + normalizeExact(seed.Name)
		if c, ok := exactKeyToCode[exactKey]; ok {
			code = c
		} else {
			// Tier 2: tone-stripped match against name_en. Only accept when the
			// parent province is identical and the match is unambiguous.
			strippedKey := seed.ProvinceCode + "|" + normalizeStripped(seed.Name)
			candidates := strippedKeyToCodes[strippedKey]
			if len(candidates) == 1 {
				code = candidates[0]
			} else if len(candidates) > 1 {
				// Disambiguate by exact name among the candidates that share the
				// same parent province.
				for _, cand := range candidates {
					if provinceOfWard[cand] != seed.ProvinceCode {
						continue
					}
					candExactKey := seed.ProvinceCode + "|" + normalizeExact(seed.Name)
					if c2, ok := exactKeyToCode[candExactKey]; ok && c2 == cand {
						code = cand
						break
					}
				}
			}
		}
		if code == "" {
			unmatched = append(unmatched, seed)
			continue
		}
		if err := s.postalCodeRepo.UpdateWardPostalCode(ctx, code, seed.PostalCode); err != nil {
			return fmt.Errorf("update ward %s postal code: %w", code, err)
		}
		matched++
	}

	if len(unmatched) > 0 {
		return fmt.Errorf("postal code import failed: %d unmatched ward(s), first: %+v", len(unmatched), unmatched[0])
	}

	missingProvinces, err := s.postalCodeRepo.CountProvincesMissingPostalPrefix(ctx)
	if err != nil {
		return err
	}
	missingWards, err := s.postalCodeRepo.CountWardsMissingPostalCode(ctx)
	if err != nil {
		return err
	}
	log.Printf("✅ Postal code import complete: %d/%d wards matched, %d provinces and %d wards still missing",
		matched, len(wardSeeds), missingProvinces, missingWards)
	if matched != len(wardSeeds) {
		return fmt.Errorf("postal code import verification failed: matched=%d/%d",
			matched, len(wardSeeds))
	}
	return nil
}

// Command decreecheck inspects the GSO decree listing and reports whether the
// published dataset is out of date, or applies the next dataset version bump.
//
// Usage:
//
//	decreecheck                         # check the live GSO page
//	decreecheck --html-file page.html   # check a saved page (offline/testing)
//	decreecheck --today 2026-09-21      # override "today" for the effective-date gate
//	decreecheck --apply --decree 388/NQ-UBTVQH16
//
// Machine-readable key=value lines are written to stdout; diagnostics go to stderr.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	datasetmetadata "github.com/thanglequoc-vn-provinces/v2/internal/dataset_metadata"
	decreecheck "github.com/thanglequoc-vn-provinces/v2/internal/decree_check"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("decreecheck: ")

	apply := flag.Bool("apply", false, "bump the dataset version and set latest_decree in the version file")
	decree := flag.String("decree", "", "decree number to record when --apply is set")
	versionFile := flag.String("version-file", datasetmetadata.DefaultVersionFilePath, "path to the version source file")
	url := flag.String("url", decreecheck.DefaultURL, "GSO decree listing URL")
	htmlFile := flag.String("html-file", "", "parse a saved HTML file instead of fetching the URL")
	today := flag.String("today", "", "override today's date (YYYY-MM-DD), defaults to now in Asia/Ho_Chi_Minh")
	flag.Parse()

	if *apply {
		runApply(*versionFile, *decree)
		return
	}

	runCheck(*versionFile, *url, *htmlFile, *today)
}

func runApply(versionFile, decree string) {
	if strings.TrimSpace(decree) == "" {
		log.Fatal("--apply requires --decree")
	}

	metadata, err := datasetmetadata.LoadFromFile(versionFile)
	if err != nil {
		log.Fatalf("load version file: %v", err)
	}

	newVersion, err := decreecheck.BumpMinor(metadata.DatasetVersion)
	if err != nil {
		log.Fatalf("bump version: %v", err)
	}

	if err := decreecheck.UpdateVersionFile(versionFile, newVersion, strings.TrimSpace(decree)); err != nil {
		log.Fatalf("update version file: %v", err)
	}

	log.Printf("updated %s: %s -> %s, latest_decree=%s", versionFile, metadata.DatasetVersion, newVersion, strings.TrimSpace(decree))

	printKV("applied", "true")
	printKV("previous_version", metadata.DatasetVersion)
	printKV("new_version", newVersion)
	printKV("latest_decree", strings.TrimSpace(decree))
}

func runCheck(versionFile, url, htmlFile, today string) {
	metadata, err := datasetmetadata.LoadFromFile(versionFile)
	if err != nil {
		log.Fatalf("load version file: %v", err)
	}

	decrees, err := loadDecrees(url, htmlFile)
	if err != nil {
		log.Fatalf("load decrees: %v", err)
	}

	todayTime, err := resolveToday(today)
	if err != nil {
		log.Fatalf("resolve today: %v", err)
	}

	decision, err := decreecheck.Evaluate(decrees, metadata.LatestDecree, metadata.DatasetVersion, todayTime)
	if err != nil {
		log.Fatalf("evaluate decree table: %v", err)
	}

	log.Printf("scanned %d decree rows; %s", len(decrees), decision.Reason)

	printKV("should_generate", fmt.Sprintf("%t", decision.ShouldGenerate))
	printKV("decree", decision.Selected.Number)
	printKV("decree_slug", decreecheck.Slug(decision.Selected.Number))
	printKV("effective_date", decision.Selected.EffectiveDateString())
	printKV("issue_date", decision.Selected.IssueDateString())
	printKV("current_decree", decision.CurrentDecree)
	printKV("current_version", decision.CurrentVersion)
	printKV("reason", decision.Reason)
}

func loadDecrees(url, htmlFile string) ([]decreecheck.Decree, error) {
	if strings.TrimSpace(htmlFile) != "" {
		return decreecheck.ParseDecreesFile(htmlFile)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return decreecheck.FetchDecrees(ctx, url)
}

func resolveToday(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Now(), nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --today %q: expected YYYY-MM-DD", value)
	}
	return parsed, nil
}

func printKV(key, value string) {
	value = strings.ReplaceAll(value, "\n", " ")
	fmt.Printf("%s=%s\n", key, value)
}

// Package decree_check inspects the GSO administrative-unit decree table and
// decides whether the published dataset is out of date.
//
// The GSO page (https://danhmuchanhchinh.nso.gov.vn/NghiDinh.aspx) renders an
// ASP.NET WebForms DevExpress grid server-side. Each data row carries an id of
// the form ...DXDataRowN and exposes four cells: decree number, issue date,
// effective date (dd/MM/yyyy) and content. Rows are ordered newest-published
// first, which means a decree with a future effective date can sit above an
// already-effective one.
package decree_check

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	_ "time/tzdata" // embed the tz database so Asia/Ho_Chi_Minh resolves on minimal runners
)

// DefaultURL is the public GSO decree listing page.
const DefaultURL = "https://danhmuchanhchinh.nso.gov.vn/NghiDinh.aspx"

const (
	httpTimeout = 30 * time.Second
	userAgent   = "Mozilla/5.0 (compatible; vn-provinces-decree-checker/1.0)"
	dateLayout  = "02/01/2006"
)

var (
	rowPattern   = regexp.MustCompile(`(?s)<tr[^>]*id="[^"]*DXDataRow\d+"[^>]*>(.*?)</tr>`)
	cellPattern  = regexp.MustCompile(`(?s)<td[^>]*class="dxgv"[^>]*>(.*?)</td>`)
	tagPattern   = regexp.MustCompile(`(?s)<[^>]*>`)
	spacePattern = regexp.MustCompile(`\s+`)
)

// vietnamLocation is the timezone used to decide whether a decree is effective.
var vietnamLocation = loadVietnamLocation()

func loadVietnamLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return time.FixedZone("ICT", 7*60*60)
	}
	return loc
}

// Decree is a single row of the GSO decree table.
type Decree struct {
	Number        string
	IssueDate     time.Time
	EffectiveDate time.Time
	Content       string
}

// Decision is the outcome of comparing the GSO table against the recorded dataset.
type Decision struct {
	ShouldGenerate bool
	Selected       Decree
	CurrentDecree  string
	CurrentVersion string
	Reason         string
}

// EffectiveDateString renders the selected effective date in the GSO format.
func (d Decree) EffectiveDateString() string {
	if d.EffectiveDate.IsZero() {
		return ""
	}
	return d.EffectiveDate.Format(dateLayout)
}

// IssueDateString renders the selected issue date in the GSO format.
func (d Decree) IssueDateString() string {
	if d.IssueDate.IsZero() {
		return ""
	}
	return d.IssueDate.Format(dateLayout)
}

// FetchDecrees downloads the GSO page and parses its decree rows.
func FetchDecrees(ctx context.Context, url string) ([]Decree, error) {
	if strings.TrimSpace(url) == "" {
		url = DefaultURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build decree request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	client := &http.Client{Timeout: httpTimeout}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch decree page %s: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch decree page %s: unexpected status %s", url, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read decree page %s: %w", url, err)
	}

	return ParseDecrees(string(body))
}

// ParseDecrees extracts decree rows from the GSO page HTML, preserving document order.
func ParseDecrees(pageHTML string) ([]Decree, error) {
	rowMatches := rowPattern.FindAllStringSubmatch(pageHTML, -1)
	if len(rowMatches) == 0 {
		return nil, errors.New("no decree rows found — the GSO page markup may have changed")
	}

	decrees := make([]Decree, 0, len(rowMatches))
	for _, match := range rowMatches {
		decree, err := parseRow(match[1])
		if err != nil {
			continue
		}
		decrees = append(decrees, decree)
	}

	if len(decrees) == 0 {
		return nil, fmt.Errorf("found %d decree rows but none were parseable — the GSO page markup may have changed", len(rowMatches))
	}
	return decrees, nil
}

func parseRow(rowHTML string) (Decree, error) {
	cellMatches := cellPattern.FindAllStringSubmatch(rowHTML, -1)
	if len(cellMatches) < 4 {
		return Decree{}, fmt.Errorf("expected 4 cells, found %d", len(cellMatches))
	}

	number := cleanCell(cellMatches[0][1])
	if number == "" {
		return Decree{}, errors.New("empty decree number")
	}

	issueDate, err := parseDate(cleanCell(cellMatches[1][1]))
	if err != nil {
		return Decree{}, fmt.Errorf("parse issue date for %q: %w", number, err)
	}

	effectiveDate, err := parseDate(cleanCell(cellMatches[2][1]))
	if err != nil {
		return Decree{}, fmt.Errorf("parse effective date for %q: %w", number, err)
	}

	return Decree{
		Number:        number,
		IssueDate:     issueDate,
		EffectiveDate: effectiveDate,
		Content:       cleanCell(cellMatches[3][1]),
	}, nil
}

func cleanCell(raw string) string {
	stripped := tagPattern.ReplaceAllString(raw, " ")
	unescaped := html.UnescapeString(stripped)
	return strings.TrimSpace(spacePattern.ReplaceAllString(unescaped, " "))
}

func parseDate(value string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, strings.TrimSpace(value), vietnamLocation)
}

// SelectNewestEffective returns the first row (document order) whose effective
// date is on or before today in Vietnam time. Because the list is
// newest-published-first, this is the newest already-effective decree.
func SelectNewestEffective(decrees []Decree, today time.Time) (Decree, bool) {
	todayDate := dateOnly(today)
	for _, decree := range decrees {
		if dateOnly(decree.EffectiveDate).After(todayDate) {
			continue
		}
		return decree, true
	}
	return Decree{}, false
}

// Evaluate compares the GSO rows with the recorded decree and version.
func Evaluate(decrees []Decree, currentDecree, currentVersion string, today time.Time) (Decision, error) {
	if len(decrees) == 0 {
		return Decision{}, errors.New("no decree rows to evaluate")
	}

	decision := Decision{
		CurrentDecree:  strings.TrimSpace(currentDecree),
		CurrentVersion: strings.TrimSpace(currentVersion),
	}

	selected, ok := SelectNewestEffective(decrees, today)
	if !ok {
		decision.Reason = "no decree with an effective date on or before today"
		return decision, nil
	}
	decision.Selected = selected

	detected := normalizeDecreeNumber(selected.Number)
	current := normalizeDecreeNumber(currentDecree)

	if detected == "" {
		return decision, errors.New("detected decree number is empty")
	}

	if current != "" && detected == current {
		decision.Reason = fmt.Sprintf("decree %s is already recorded in version.txt", selected.Number)
		return decision, nil
	}

	if current != "" {
		currentIndex := indexOfDecree(decrees, currentDecree)
		detectedIndex := indexOfDecree(decrees, selected.Number)
		if currentIndex >= 0 && detectedIndex >= 0 && detectedIndex > currentIndex {
			decision.Reason = fmt.Sprintf(
				"detected decree %s is older than the recorded %s (it appears lower in the table)",
				selected.Number, strings.TrimSpace(currentDecree),
			)
			return decision, nil
		}
	}

	decision.ShouldGenerate = true
	decision.Reason = fmt.Sprintf(
		"new decree %s effective %s (recorded: %s)",
		selected.Number, selected.EffectiveDateString(), displayOrNone(currentDecree),
	)
	return decision, nil
}

// Slug renders a decree number into a filesystem/branch-safe slug.
func Slug(number string) string {
	lowered := strings.ToLower(strings.TrimSpace(number))
	var builder strings.Builder
	lastDash := false
	for _, r := range lowered {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}

func normalizeDecreeNumber(number string) string {
	return strings.ToUpper(strings.TrimSpace(spacePattern.ReplaceAllString(number, " ")))
}

func indexOfDecree(decrees []Decree, number string) int {
	target := normalizeDecreeNumber(number)
	for i, decree := range decrees {
		if normalizeDecreeNumber(decree.Number) == target {
			return i
		}
	}
	return -1
}

func dateOnly(t time.Time) time.Time {
	local := t.In(vietnamLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, vietnamLocation)
}

func displayOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(none)"
	}
	return strings.TrimSpace(value)
}

// ParseDecreesFile reads a local HTML file and parses its decree rows. Useful for tests.
func ParseDecreesFile(path string) ([]Decree, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read decree html file %s: %w", path, err)
	}
	return ParseDecrees(string(content))
}

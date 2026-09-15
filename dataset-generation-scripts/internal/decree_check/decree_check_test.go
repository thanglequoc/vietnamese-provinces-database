package decree_check

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixturePath = "testdata/nghidinh_sample.html"

func mustLoadFixture(t *testing.T) []Decree {
	t.Helper()
	decrees, err := ParseDecreesFile(fixturePath)
	require.NoError(t, err)
	require.NotEmpty(t, decrees)
	return decrees
}

func date(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02", value, vietnamLocation)
	require.NoError(t, err)
	return parsed
}

func TestParseDecreesFromFixture(t *testing.T) {
	decrees := mustLoadFixture(t)

	require.Len(t, decrees, 6)

	first := decrees[0]
	assert.Equal(t, "388/NQ-UBTVQH16", first.Number)
	assert.Equal(t, "04/08/2026", first.IssueDateString())
	assert.Equal(t, "20/09/2026", first.EffectiveDateString())
	assert.Contains(t, first.Content, "thành lập các phường thuộc tỉnh Bắc Ninh")

	// Dates must be anchored in Vietnam time.
	assert.Equal(t, vietnamLocation, first.EffectiveDate.Location())

	assert.Equal(t, "36/2026/QH16", decrees[2].Number)
	assert.Equal(t, "01/09/2026", decrees[2].EffectiveDateString())
}

func TestParseDecreesUnescapesAndTrimsCells(t *testing.T) {
	page := `<table>` +
		`<tr id="ctl00_PlaceHolderMain_ASPxGridView1_DXDataRow0" class="dxgvDataRow_Office2003_Blue">` +
		`<td class="dxgv">1279/NQ-UBTVQH15 </td>` +
		`<td class="dxgv">14/11/2024</td>` +
		`<td class="dxgv">01/01/2025</td>` +
		`<td class="dxgv">Ngh&#7883; quy&#7871;t &amp; test</td>` +
		`</tr></table>`

	decrees, err := ParseDecrees(page)
	require.NoError(t, err)
	require.Len(t, decrees, 1)

	assert.Equal(t, "1279/NQ-UBTVQH15", decrees[0].Number)
	assert.Equal(t, "Nghị quyết & test", decrees[0].Content)
}

func TestParseDecreesErrorsWhenMarkupChanges(t *testing.T) {
	_, err := ParseDecrees("<html><body>no grid here</body></html>")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "markup may have changed")
}

func TestSelectNewestEffective(t *testing.T) {
	decrees := mustLoadFixture(t)

	selected, ok := SelectNewestEffective(decrees, date(t, "2026-09-16"))
	require.True(t, ok)
	assert.Equal(t, "36/2026/QH16", selected.Number)

	selected, ok = SelectNewestEffective(decrees, date(t, "2026-09-21"))
	require.True(t, ok)
	assert.Equal(t, "388/NQ-UBTVQH16", selected.Number)

	selected, ok = SelectNewestEffective(decrees, date(t, "2026-09-01"))
	require.True(t, ok)
	assert.Equal(t, "36/2026/QH16", selected.Number)
}

func TestSelectNewestEffectiveNoneEffectiveYet(t *testing.T) {
	decrees := mustLoadFixture(t)

	_, ok := SelectNewestEffective(decrees, date(t, "2020-01-01"))
	assert.False(t, ok)
}

func TestEvaluateNoOpWhenRecordedDecreeAlreadyEffective(t *testing.T) {
	decrees := mustLoadFixture(t)

	decision, err := Evaluate(decrees, "36/2026/QH16", "v5.1.0", date(t, "2026-09-16"))
	require.NoError(t, err)

	assert.False(t, decision.ShouldGenerate)
	assert.Equal(t, "36/2026/QH16", decision.Selected.Number)
	assert.Contains(t, decision.Reason, "already recorded")
}

func TestEvaluateDetectsNewlyEffectiveDecree(t *testing.T) {
	decrees := mustLoadFixture(t)

	decision, err := Evaluate(decrees, "36/2026/QH16", "v5.1.0", date(t, "2026-09-21"))
	require.NoError(t, err)

	assert.True(t, decision.ShouldGenerate)
	assert.Equal(t, "388/NQ-UBTVQH16", decision.Selected.Number)
	assert.Equal(t, "20/09/2026", decision.Selected.EffectiveDateString())
	assert.Equal(t, "36/2026/QH16", decision.CurrentDecree)
	assert.Equal(t, "v5.1.0", decision.CurrentVersion)
}

func TestEvaluateSkipsFutureEffectiveDecree(t *testing.T) {
	decrees := mustLoadFixture(t)

	// 388/NQ and 39/2026 are effective 20/09/2026, so they are not actionable yet.
	decision, err := Evaluate(decrees, "30/2026/QH16", "v5.1.0", date(t, "2026-09-01"))
	require.NoError(t, err)

	assert.True(t, decision.ShouldGenerate)
	assert.Equal(t, "36/2026/QH16", decision.Selected.Number)
}

func TestEvaluateOrderingGuardPreventsDowngrade(t *testing.T) {
	decrees := mustLoadFixture(t)

	// Recorded decree sits at the top of the table but is not yet effective.
	// The scanner must not fall back to an older already-effective decree.
	decision, err := Evaluate(decrees, "388/NQ-UBTVQH16", "v5.1.0", date(t, "2026-09-16"))
	require.NoError(t, err)

	assert.False(t, decision.ShouldGenerate)
	assert.Contains(t, decision.Reason, "older than the recorded")
}

func TestEvaluateWithNoRecordedDecree(t *testing.T) {
	decrees := mustLoadFixture(t)

	decision, err := Evaluate(decrees, "", "v5.1.0", date(t, "2026-09-21"))
	require.NoError(t, err)

	assert.True(t, decision.ShouldGenerate)
	assert.Equal(t, "388/NQ-UBTVQH16", decision.Selected.Number)
	assert.Contains(t, decision.Reason, "(none)")
}

func TestEvaluateErrorsOnEmptyTable(t *testing.T) {
	_, err := Evaluate(nil, "36/2026/QH16", "v5.1.0", date(t, "2026-09-21"))
	require.Error(t, err)
}

func TestSlug(t *testing.T) {
	assert.Equal(t, "388-nq-ubtvqh16", Slug("388/NQ-UBTVQH16"))
	assert.Equal(t, "74-bocaphuyen-nq-cp", Slug("74-BOCAPHUYEN/NQ-CP"))
	assert.Equal(t, "36-2026-qh16", Slug(" 36/2026/QH16 "))
}

package analytics

import "testing"

// TestMaxRangeDays — sanity-check tier mapping.
// Free=7, Pro/Pro_yearly=90, Max/Max_yearly=365, unknown → 7 (safe fallback).
func TestMaxRangeDays(t *testing.T) {
	cases := []struct {
		plan string
		want int
	}{
		{"free", 7},
		{"pro", 90},
		{"pro_yearly", 90},
		{"max", 365},
		{"max_yearly", 365},
		// Whitelist: любой неизвестный plan_id → free (safe default).
		// Раньше strings.HasPrefix ошибочно маппил "professional"→pro, "maximum"→max.
		{"professional", 7},
		{"proto", 7},
		{"maximum", 7},
		{"", 7},
	}
	for _, tc := range cases {
		t.Run(tc.plan, func(t *testing.T) {
			got := MaxRangeDays(tc.plan)
			if got != tc.want {
				t.Errorf("MaxRangeDays(%q) = %d, want %d", tc.plan, got, tc.want)
			}
		})
	}
}

// TestClampRange покрывает 4 тира × 4 запрошенных диапазона = 16 комбинаций
// + несколько edge-cases (unknown tier, unknown range → fallback).
func TestClampRange(t *testing.T) {
	cases := []struct {
		name      string
		requested RangeID
		plan      string
		want      RangeID
	}{
		// Free (7d cap)
		{"free_7d_stays", Range7d, "free", Range7d},
		{"free_30d_clamped", Range30d, "free", Range7d},
		{"free_90d_clamped", Range90d, "free", Range7d},
		{"free_365d_clamped", Range365d, "free", Range7d},

		// Pro (90d cap)
		{"pro_7d_stays", Range7d, "pro", Range7d},
		{"pro_30d_stays", Range30d, "pro", Range30d},
		{"pro_90d_stays", Range90d, "pro", Range90d},
		{"pro_365d_clamped", Range365d, "pro", Range90d},

		// Pro yearly — та же семантика
		{"pro_yearly_365d_clamped", Range365d, "pro_yearly", Range90d},

		// Max (365d cap — ничего не обрезается)
		{"max_7d_stays", Range7d, "max", Range7d},
		{"max_30d_stays", Range30d, "max", Range30d},
		{"max_90d_stays", Range90d, "max", Range90d},
		{"max_365d_stays", Range365d, "max", Range365d},
		{"max_yearly_365d_stays", Range365d, "max_yearly", Range365d},

		// Unknown plan → treated as free (H3 защита).
		{"unknown_plan_365d_clamped_to_7d", Range365d, "professional", Range7d},
		{"unknown_plan_7d_stays", Range7d, "maximum", Range7d},

		// Unknown range — rangeToDays возвращает 7, клампа не требуется.
		{"unknown_range_free", RangeID("forever"), "free", RangeID("forever")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClampRange(tc.requested, tc.plan)
			if got != tc.want {
				t.Errorf("ClampRange(%q, %q) = %q, want %q", tc.requested, tc.plan, got, tc.want)
			}
		})
	}
}

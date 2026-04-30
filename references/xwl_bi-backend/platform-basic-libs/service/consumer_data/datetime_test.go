package consumer_data

import (
	"testing"

	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

func TestNormalizeDateTimeForClickHouse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{
			name:  "full timestamp stays unchanged",
			input: "2026-04-08 16:14:53",
			want:  "2026-04-08 16:14:53",
		},
		{
			name:  "date only becomes local midnight",
			input: "2026-04-08",
			want:  "2026-04-08 00:00:00",
		},
		{
			name:      "invalid format is rejected",
			input:     "2026/04/08",
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeDateTimeForClickHouse(tc.input)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error for %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeDateTimeForClickHouse(%q) returned error: %v", tc.input, err)
			}
			if got.Format(util.TimeFormat) != tc.want {
				t.Fatalf("normalizeDateTimeForClickHouse(%q) = %q, want %q", tc.input, got.Format(util.TimeFormat), tc.want)
			}
		})
	}
}

func TestNormalizeReportTime(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "完整时间保持不变", input: "2026-04-08 16:14:53", want: "2026-04-08 16:14:53"},
		{name: "仅日期补齐到零点", input: "2026-04-08", want: "2026-04-08 00:00:00"},
		{name: "无法识别时保留原值", input: "2026/04/08", want: "2026/04/08"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeReportTime(tc.input); got != tc.want {
				t.Fatalf("NormalizeReportTime(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeClientTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{name: "完整时间可解析", input: "2026-04-08 16:14:53", want: "2026-04-08 16:14:53"},
		{name: "仅日期补齐后可解析", input: "2026-04-08", want: "2026-04-08 00:00:00"},
		{name: "非法时间返回错误", input: "2026/04/08", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, parsed, err := NormalizeClientTime(tc.input)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error for %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeClientTime(%q) returned error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeClientTime(%q) = %q, want %q", tc.input, got, tc.want)
			}
			if parsed.Format(util.TimeFormat) != tc.want {
				t.Fatalf("parsed time = %q, want %q", parsed.Format(util.TimeFormat), tc.want)
			}
		})
	}
}

func TestReportTimeHasClock(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "完整时间保留时分秒语义", input: "2026-04-08 16:14:53", want: true},
		{name: "仅日期不算带时分秒", input: "2026-04-08", want: false},
		{name: "非法时间返回 false", input: "bad-time", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ReportTimeHasClock(tc.input); got != tc.want {
				t.Fatalf("ReportTimeHasClock(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeReportTimeForValidation(t *testing.T) {
	tests := []struct {
		name         string
		reportTime   string
		fallback     string
		want         string
		wantHasClock bool
	}{
		{name: "完整时间保留时分秒", reportTime: "2026-04-08 16:14:53", fallback: "2026-04-08 10:11:12", want: "2026-04-08 16:14:53", wantHasClock: true},
		{name: "仅日期归一化但不算带时分秒", reportTime: "2026-04-08", fallback: "2026-04-08 10:11:12", want: "2026-04-08 00:00:00", wantHasClock: false},
		{name: "无法解析时回退到 fallback", reportTime: "bad-time", fallback: "2026-04-08 10:11:12", want: "2026-04-08 10:11:12", wantHasClock: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, hasClock := NormalizeReportTimeForValidation(tc.reportTime, tc.fallback)
			if got != tc.want {
				t.Fatalf("NormalizeReportTimeForValidation(%q, %q) = %q, want %q", tc.reportTime, tc.fallback, got, tc.want)
			}
			if hasClock != tc.wantHasClock {
				t.Fatalf("hasClock = %v, want %v", hasClock, tc.wantHasClock)
			}
		})
	}
}

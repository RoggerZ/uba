package controller

import "testing"

func TestApplyDebugClockSkewCheck(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		reportTime   string
		wantFail     bool
		wantReason   bool
		wantDataType string
	}{
		{
			name:         "clock skew over ten minutes fails validation",
			body:         []byte(`{"xwl_update_time":"2026-04-09 10:00:00"}`),
			reportTime:   "2026-04-09 10:11:00",
			wantFail:     true,
			wantReason:   true,
			wantDataType: "事件属性类型不合法",
		},
		{
			name:         "clock skew within ten minutes passes",
			body:         []byte(`{"xwl_update_time":"2026-04-09 10:00:00"}`),
			reportTime:   "2026-04-09 10:09:00",
			wantFail:     false,
			wantReason:   false,
			wantDataType: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := map[string]interface{}{}
			haveFailAttr := false

			applyDebugClockSkewCheck(tc.body, tc.reportTime, "事件属性类型不合法", m, &haveFailAttr)

			if haveFailAttr != tc.wantFail {
				t.Fatalf("haveFailAttr = %v, want %v", haveFailAttr, tc.wantFail)
			}

			_, hasReason := m["error_reason"]
			if hasReason != tc.wantReason {
				t.Fatalf("has error_reason = %v, want %v", hasReason, tc.wantReason)
			}

			gotDataType, _ := m["data_judge"].(string)
			if gotDataType != tc.wantDataType {
				t.Fatalf("data_judge = %q, want %q", gotDataType, tc.wantDataType)
			}
		})
	}
}

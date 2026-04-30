package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticAssetsUseLongCacheHeaders(t *testing.T) {
	app := Init()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "本地 vendor",
			path: "/vendor/ajax/libs/echarts/4.7.0/echarts.min.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatalf("request static asset: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}

			const wantCacheControl = "public, max-age=2592000"
			if got := resp.Header.Get("Cache-Control"); got != wantCacheControl {
				t.Fatalf("Cache-Control = %q, want %q", got, wantCacheControl)
			}
		})
	}
}

func TestIndexDoesNotUseLongStaticCache(t *testing.T) {
	app := Init()

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request index: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if got := resp.Header.Get("Cache-Control"); got == "public, max-age=2592000" {
		t.Fatalf("index.html should not use long static cache, got %q", got)
	}
}

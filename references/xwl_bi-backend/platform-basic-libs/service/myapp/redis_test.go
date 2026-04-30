package myapp

import "testing"

func TestBuildAppTableIDKey(t *testing.T) {
	if got := BuildAppTableIDKey("1001", "demo"); got != "1001_xwl_demo" {
		t.Fatalf("BuildAppTableIDKey() = %q, want %q", got, "1001_xwl_demo")
	}
}

func TestNotifyAppTableIDCacheChanged(t *testing.T) {
	var calledWith string
	RegisterAppTableIDCacheInvalidator(func(cacheKey string) {
		calledWith = cacheKey
	})

	NotifyAppTableIDCacheChanged("1001_xwl_demo")
	if calledWith != "1001_xwl_demo" {
		t.Fatalf("calledWith = %q, want %q", calledWith, "1001_xwl_demo")
	}
}

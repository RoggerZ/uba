package application

import "testing"

func TestBuildClickHouseStdDSN(t *testing.T) {
	got := buildClickHouseStdDSN("127.0.0.1", "9000", "xwl_bi", "user", "pass", 12345)
	want := "clickhouse://127.0.0.1:9000/xwl_bi?compress=lz4&max_query_size=12345&password=pass&username=user"
	if got != want {
		t.Fatalf("buildClickHouseStdDSN() = %q, want %q", got, want)
	}
}

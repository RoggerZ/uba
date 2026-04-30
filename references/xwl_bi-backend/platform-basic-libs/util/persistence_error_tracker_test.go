package util

import (
	"errors"
	"testing"
)

func TestClassifyPersistenceErrorSchemaChangeDoesNotCountTowardCircuitBreak(t *testing.T) {
	classification := ClassifyPersistenceError("clickhouse_schema_change_failed", errors.New("permission denied"))
	if classification.ErrorClass != "clickhouse_schema_change_failed" {
		t.Fatalf("ErrorClass = %q, want clickhouse_schema_change_failed", classification.ErrorClass)
	}
	if classification.CountTowardCircuitBreak {
		t.Fatal("schema change failure should not count toward circuit break")
	}
}

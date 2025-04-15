// clickhouse_test.go
package storage

import (
	"testing"
)

func TestClickHouseConnect(t *testing.T) {
	service := NewClickHouseService()
	err := service.Connect("localhost", "9000", "default", "default", "test-jwt-token", false)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer service.db.Close()
}

func TestImportData(t *testing.T) {
	service := NewClickHouseService()
	err := service.Connect("localhost", "9000", "default", "default", "test-jwt-token", false)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer service.db.Close()

	data := []map[string]interface{}{
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25},
	}

	count, err := service.ImportData("test_table", []string{"name", "age"}, data)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}
}
package helpers

import (
	"testing"
	"time"
)

func TestDateAndTime(t *testing.T) {
	mockHexString := "7FC76122"

	expectedTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(0x2261C77F * time.Second)

	result := DateAndTime(mockHexString)

	if result == nil {
		t.Fatalf("expected non-nil result, got nil")
	}

	finalTime, ok := result.(time.Time)
	if !ok {
		t.Fatalf("expected result to be of type time.Time, got %T", result)
	}

	if !finalTime.Equal(expectedTime) {
		t.Errorf("expected %v, got %v", expectedTime, finalTime)
	}
}
func TestParseDatetime(t *testing.T) {
	mockHexString := "240830011917"

	expectedDatetime := "2024-08-30T01:19:17Z"

	result, err := ParseDatetime(mockHexString)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expectedDatetime {
		t.Errorf("expected %s, got %s", expectedDatetime, result)
	}

	invalidHexString := "1234"
	_, err = ParseDatetime(invalidHexString)
	if err == nil {
		t.Fatalf("expected error for invalid input length, got nil")
	}
}

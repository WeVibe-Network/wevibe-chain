package cmd

import "testing"

func TestValidateEpochDurationSecondsEnv(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		t.Setenv(epochDurationSecondsEnv, "")
		if err := validateEpochDurationSecondsEnv(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("valid positive integer", func(t *testing.T) {
		t.Setenv(epochDurationSecondsEnv, "2")
		if err := validateEpochDurationSecondsEnv(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("zero rejected", func(t *testing.T) {
		t.Setenv(epochDurationSecondsEnv, "0")
		if err := validateEpochDurationSecondsEnv(); err == nil {
			t.Fatal("expected error for zero value")
		}
	})

	t.Run("non numeric rejected", func(t *testing.T) {
		t.Setenv(epochDurationSecondsEnv, "abc")
		if err := validateEpochDurationSecondsEnv(); err == nil {
			t.Fatal("expected error for non numeric value")
		}
	})
}

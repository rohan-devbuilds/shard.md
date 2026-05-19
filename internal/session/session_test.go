package session

import "testing"

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens(""); got != 0 {
		t.Fatalf("EstimateTokens(empty) = %d, want 0", got)
	}
	if got := EstimateTokens("12345678"); got != 2 {
		t.Fatalf("EstimateTokens(8 chars) = %d, want 2", got)
	}
}

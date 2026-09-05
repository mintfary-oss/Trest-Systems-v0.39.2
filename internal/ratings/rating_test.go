package ratings

import "testing"

func TestValidateScore(t *testing.T) {
	for _, score := range []float64{1, 3.5, 5} {
		if err := ValidateScore(score); err != nil {
			t.Fatalf("score %v rejected: %v", score, err)
		}
	}
	for _, score := range []float64{0, 5.01, 10} {
		if err := ValidateScore(score); err == nil {
			t.Fatalf("score %v accepted", score)
		}
	}
}
func TestValidateTarget(t *testing.T) {
	if ValidateTarget("contractor") != nil || ValidateTarget("supplier") != nil {
		t.Fatal("valid targets rejected")
	}
	if ValidateTarget("customer") == nil {
		t.Fatal("invalid target accepted")
	}
}
func TestEligibleOrderStatus(t *testing.T) {
	if !EligibleOrderStatus("completed") {
		t.Fatal("completed must be eligible")
	}
	if EligibleOrderStatus("in_progress") {
		t.Fatal("in_progress must not be eligible")
	}
}

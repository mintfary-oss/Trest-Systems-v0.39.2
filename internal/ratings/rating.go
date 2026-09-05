package ratings

import "fmt"

func ValidateScore(score float64) error {
	if score < 1 || score > 5 {
		return fmt.Errorf("score must be between 1 and 5")
	}
	return nil
}
func ValidateTarget(targetType string) error {
	if targetType != "contractor" && targetType != "supplier" {
		return fmt.Errorf("target_type must be contractor or supplier")
	}
	return nil
}
func EligibleOrderStatus(status string) bool { return status == "completed" }

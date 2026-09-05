package suppliers

func Eligible(verificationStatus string, active bool, category, requiredCategory, region string, supportedRegions []string, stock, quantity float64, published bool) bool {
	if verificationStatus != "verified" || !active || !published || stock < quantity || quantity <= 0 {
		return false
	}
	if requiredCategory != "" && category != requiredCategory {
		return false
	}
	if region != "" && !contains(supportedRegions, region) {
		return false
	}
	return true
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

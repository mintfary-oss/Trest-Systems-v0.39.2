package contractors

// Eligible reports whether a contractor satisfies the deterministic baseline
// used by matching: verified+active status, requested competency and geography.
func Eligible(verificationStatus string, active bool, contractorCompetencies, requiredCompetencies, contractorRegions, requiredRegions []string) bool {
	if verificationStatus != "verified" || !active {
		return false
	}
	if !containsAny(contractorCompetencies, requiredCompetencies) && len(requiredCompetencies) > 0 {
		return false
	}
	if len(requiredRegions) > 0 && !containsAny(contractorRegions, requiredRegions) {
		return false
	}
	return true
}

func containsAny(have, want []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, v := range have {
		set[v] = struct{}{}
	}
	for _, v := range want {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

package suppliers

import "testing"

func TestEligible(t *testing.T) {
	regions := []string{"ru", "by"}
	cases := []struct {
		name, status               string
		active, published          bool
		category, required, region string
		stock, qty                 float64
		want                       bool
	}{
		{"ok", "verified", true, true, "cement", "cement", "ru", 100, 20, true}, {"unverified", "pending", true, true, "cement", "cement", "ru", 100, 20, false}, {"inactive", "verified", false, true, "cement", "cement", "ru", 100, 20, false}, {"draft", "verified", true, false, "cement", "cement", "ru", 100, 20, false}, {"wrong category", "verified", true, true, "brick", "cement", "ru", 100, 20, false}, {"wrong region", "verified", true, true, "cement", "cement", "kz", 100, 20, false}, {"not enough stock", "verified", true, true, "cement", "cement", "ru", 10, 20, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Eligible(tc.status, tc.active, tc.category, tc.required, tc.region, regions, tc.stock, tc.qty, tc.published); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

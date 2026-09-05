package contractors

import "testing"

func TestEligible(t *testing.T) {
	cases := []struct {
		name           string
		status         string
		active         bool
		cc, rc, cr, rr []string
		want           bool
	}{
		{"verified match", "verified", true, []string{"brickwork"}, []string{"brickwork"}, []string{"winnipeg"}, []string{"winnipeg"}, true},
		{"unverified", "pending", true, []string{"brickwork"}, []string{"brickwork"}, nil, nil, false},
		{"inactive", "verified", false, []string{"brickwork"}, []string{"brickwork"}, nil, nil, false},
		{"wrong skill", "verified", true, []string{"roofing"}, []string{"brickwork"}, nil, nil, false},
		{"wrong region", "verified", true, []string{"brickwork"}, []string{"brickwork"}, []string{"toronto"}, []string{"winnipeg"}, false},
		{"no filters", "verified", true, nil, nil, nil, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Eligible(tc.status, tc.active, tc.cc, tc.rc, tc.cr, tc.rr); got != tc.want {
				t.Fatalf("Eligible()=%v, want %v", got, tc.want)
			}
		})
	}
}

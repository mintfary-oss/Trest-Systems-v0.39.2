package orders

import "fmt"

const (
	Draft        = "draft"
	Published    = "published"
	Matching     = "matching"
	Contracted   = "contracted"
	InProgress   = "in_progress"
	QualityCheck = "quality_check"
	Completed    = "completed"
	Cancelled    = "cancelled"
)

var transitions = map[string]map[string]bool{
	Draft:        {Published: true, Cancelled: true},
	Published:    {Matching: true, Contracted: true, Cancelled: true},
	Matching:     {Contracted: true, Cancelled: true},
	Contracted:   {InProgress: true, Cancelled: true},
	InProgress:   {QualityCheck: true, Cancelled: true},
	QualityCheck: {Completed: true, InProgress: true},
	Completed:    {},
	Cancelled:    {},
}

func CanTransition(from, to string) bool {
	return transitions[from][to]
}

func ValidateTransition(from, to string) error {
	if from == to {
		return nil
	}
	if _, ok := transitions[from]; !ok {
		return fmt.Errorf("unknown order status: %s", from)
	}
	if !transitions[from][to] {
		return fmt.Errorf("invalid order transition: %s -> %s", from, to)
	}
	return nil
}

package usecase

import "time"

type TransportType string

func CalculateDeadline(transport string, now time.Time) time.Time {
	switch transport {
	case "scooter":
		return now.Add(15 * time.Minute)
	case "car":
		return now.Add(5 * time.Minute)
	default:
		return now.Add(30 * time.Minute)
	}
}

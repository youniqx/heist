package core

import (
	"encoding/json"
	"fmt"
	"time"
)

type VaultTTL struct {
	TTL time.Duration
}

const (
	hoursInDay = 24
	daysInWeek = 7
	daysInYear = 365
	// Year is the number of hours per year.
	Year = daysInYear * hoursInDay * time.Hour
	// Week is the number of hours per week.
	Week = daysInWeek * hoursInDay * time.Hour
	// Day is the number of hours per day.
	Day = hoursInDay * time.Hour
	// Hour is one hour.
	Hour = time.Hour
	// Minute is one minute.
	Minute = time.Minute
	// Second is one second.
	Second = time.Second
)

func NewTTL(ttl time.Duration) *VaultTTL {
	return &VaultTTL{TTL: ttl}
}

func (v *VaultTTL) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(int(v.TTL.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to format TTL %v: %w", v, err)
	}

	return data, nil
}

func (v *VaultTTL) UnmarshalJSON(bytes []byte) error {
	var seconds int
	if err := json.Unmarshal(bytes, &seconds); err != nil {
		return fmt.Errorf("failed to parse TTL %s: %w", string(bytes), err)
	}

	v.TTL = time.Duration(seconds) * Second

	return nil
}

package tracker

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return fmt.Errorf("id %q may only contain lowercase letters, numbers, dashes, and underscores", id)
	}
	return nil
}

func NewEventID(habitID string, t time.Time) string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%s", t.UTC().Format("20060102T150405000000000Z"), habitID)
	}
	return fmt.Sprintf("%s-%s-%s", t.UTC().Format("20060102T150405000000000Z"), habitID, hex.EncodeToString(b[:]))
}

func NormalizeDays(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	days := make([]string, 0, len(parts))
	for _, part := range parts {
		day := strings.ToLower(strings.TrimSpace(part))
		if day != "" {
			days = append(days, day)
		}
	}
	return days
}

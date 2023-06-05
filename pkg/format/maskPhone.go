package format

import (
	"fmt"
	"strings"
	"unicode"
)

type PhoneMaskConfig struct {
	SkipBeforeMask int
	SkipAfterMask  int
}

func MaskPhoneNumber(phone string, cfg PhoneMaskConfig) string {
	phoneR := []rune(phone)
	prefix := ""

	if strings.HasPrefix(phone, "+") {
		prefix = "+"
		phoneWithoutPref := strings.Trim(phone, "+")
		phoneR = []rune(phoneWithoutPref)
	}

	var digits []rune

	for i := range phoneR {
		if unicode.IsDigit(phoneR[i]) {
			digits = append(digits, phoneR[i])
		}
	}

	result := ""

	for i := range digits {
		if i < cfg.SkipBeforeMask || i >= len(digits)-cfg.SkipAfterMask {
			result += string(digits[i])
		} else {
			result += "*"
		}
	}

	return fmt.Sprintf("%s%s", prefix, result)
}

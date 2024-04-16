package main

import (
	"fmt"
	"regexp"
	"strings"
)

func isValidCreditCard(cardNumber string) bool {
	// Regex pattern to check starting with digit and follows the format
	pattern := `^[4-6](\d{3}-?\d{4}-?\d{4}-?\d{4})$`
	re := regexp.MustCompile(pattern)

	// Removing hyphens in the number
	cleaned := strings.ReplaceAll(cardNumber, "-", "")
	if !re.MatchString(cardNumber) || len(cleaned) != 16 {
		return false
	}

	// Check for identical consecutive digits
	for i := 0; i <= len(cleaned)-4; i++ {
		if cleaned[i] == cleaned[i+1] && cleaned[i+1] == cleaned[i+2] && cleaned[i+2] == cleaned[i+3] {
			return false
		}
	}

	return true
}

func main() {
	cardNumbers := []string{
		"4253625879615786",
		"4424424424442444",
		"5122-2368-7954-3214",
		"4123456789123456",
		"5123-4567-8912-3456",
		"4123356789123456",
		"42536258796157867",         // Invalid length
		"4424444424442444",          // Too many repeated digits
		"61234-567-8912-3456",       // card number is not divided into equal groups of 4
		"5133-3367-8912-3456",       // consecutive digits  is repeating  times.
		"5123 - 3567 - 8912 - 3456", // space '  ' and - are used as separators.
	}

	for _, card := range cardNumbers {
		fmt.Printf("Card Number: %s -> Valid: %t\n", card, isValidCreditCard(card))
	}
}

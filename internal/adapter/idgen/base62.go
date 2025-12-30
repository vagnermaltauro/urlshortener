package idgen

import (
	"errors"
	"strings"
)

// Base62 character set (0-9, A-Z, a-z) = 62 characters
// Ordered for optimal encoding: digits first, then uppercase, then lowercase
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Encode converts a 64-bit integer ID to a 7-character Base62 string
// This creates the short code used in URLs (e.g., "0Ab3XyZ")
// With 7 characters, we can represent 62^7 = 3.5 trillion unique IDs
func (g *SnowflakeGenerator) Encode(id int64) string {
	if id == 0 {
		return "0000000" // Special case: zero is "0000000"
	}

	// Convert to Base62
	chars := make([]byte, 0, 11) // Maximum 11 chars for 64-bit int
	num := uint64(id)

	for num > 0 {
		remainder := num % 62
		chars = append(chars, base62Chars[remainder])
		num /= 62
	}

	// Reverse the characters (they were built backwards)
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	// Pad with leading zeros to ensure exactly 7 characters
	// This gives consistent URL length and looks cleaner
	for len(chars) < 7 {
		chars = append([]byte{'0'}, chars...)
	}

	return string(chars)
}

// Decode converts a 7-character Base62 string back to a 64-bit integer ID
// Returns an error if the string is invalid
func (g *SnowflakeGenerator) Decode(code string) (int64, error) {
	// Validate length
	if len(code) != 7 {
		return 0, errors.New("invalid code length: must be exactly 7 characters")
	}

	var id int64
	for _, char := range code {
		// Find the value of this character in Base62
		var value int64
		switch {
		case char >= '0' && char <= '9':
			value = int64(char - '0') // 0-9
		case char >= 'A' && char <= 'Z':
			value = int64(char-'A') + 10 // 10-35
		case char >= 'a' && char <= 'z':
			value = int64(char-'a') + 36 // 36-61
		default:
			return 0, errors.New("invalid character in code: must be 0-9, A-Z, or a-z")
		}

		// Multiply current ID by 62 and add this digit
		id = id*62 + value
	}

	return id, nil
}

// isValidBase62 checks if a string contains only valid Base62 characters
func isValidBase62(s string) bool {
	for _, char := range s {
		if !strings.ContainsRune(base62Chars, char) {
			return false
		}
	}
	return true
}

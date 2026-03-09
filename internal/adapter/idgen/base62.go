package idgen

import (
	"errors"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func (g *SnowflakeGenerator) Encode(id int64) string {
	if id == 0 {
		return "0000000"
	}

	chars := make([]byte, 0, 11)
	num := uint64(id)

	for num > 0 {
		remainder := num % 62
		chars = append(chars, base62Chars[remainder])
		num /= 62
	}

	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	for len(chars) < 7 {
		chars = append([]byte{'0'}, chars...)
	}

	return string(chars)
}

func (g *SnowflakeGenerator) Decode(code string) (int64, error) {

	if len(code) != 7 {
		return 0, errors.New("invalid code length: must be exactly 7 characters")
	}

	var id int64
	for _, char := range code {

		var value int64
		switch {
		case char >= '0' && char <= '9':
			value = int64(char - '0')
		case char >= 'A' && char <= 'Z':
			value = int64(char-'A') + 10
		case char >= 'a' && char <= 'z':
			value = int64(char-'a') + 36
		default:
			return 0, errors.New("invalid character in code: must be 0-9, A-Z, or a-z")
		}

		id = id*62 + value
	}

	return id, nil
}

func isValidBase62(s string) bool {
	for _, char := range s {
		if !strings.ContainsRune(base62Chars, char) {
			return false
		}
	}
	return true
}

package utils

import (
	crypto_rand "crypto/rand"
	"math/rand"
	"strings"
)



// sanitizeUsername converts username to valid MySQL username
func UtilSanitizeName(name string, prefix string) string {
	if prefix == "" {
		prefix = "x_"
	}
	// Replace invalid characters with underscores
	invalidChars := []string{" ", "&", "?", "=", "#", "@", "!", "$", "%", "^", "*", "(", ")", "+", "[", "]", "{", "}", "|", ";", "'", "\"", "<", ">", ",", ":", "/", "\\"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Remove multiple consecutive underscores
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}

	// Trim underscores from start and end
	name = strings.Trim(name, "_")

	// Limit length (MySQL username limit is 32 characters)
	if len(name) > 20 {
		name = name[:20]
	}

	name = AddRandomToString(name, 10)

	return prefix + name
}

func randomAlphanumeric(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	if _, err := crypto_rand.Read(b); err != nil {
		// Fallback to math/rand if crypto fails
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}
		return string(b)
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

func AddRandomToString(str string, length int) string {
	return str + "_" + randomAlphanumeric(length)
}

func GenerateSecurePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	bytes := make([]byte, length)
	if _, err := crypto_rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

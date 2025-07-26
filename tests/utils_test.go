package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Test time formatting and validation
func TestTimeHandling(t *testing.T) {
	t.Run("Time Formatting", func(t *testing.T) {
		now := time.Now()
		formatted := now.Format(time.RFC3339)

		parsed, err := time.Parse(time.RFC3339, formatted)
		if err != nil {
			t.Errorf("Failed to parse formatted time: %v", err)
		}

		if parsed.Unix() != now.Unix() {
			t.Error("Parsed time should match original time")
		}
	})

	t.Run("Session Duration", func(t *testing.T) {
		duration := 24 * time.Hour * 7 // 7 days
		if duration.Hours() != 168 {
			t.Errorf("Expected 168 hours, got %f", duration.Hours())
		}
	})

	t.Run("Time Comparison", func(t *testing.T) {
		now := time.Now()
		future := now.Add(1 * time.Hour)
		past := now.Add(-1 * time.Hour)

		if !future.After(now) {
			t.Error("Future time should be after current time")
		}

		if !past.Before(now) {
			t.Error("Past time should be before current time")
		}
	})

	t.Run("Time Zone Handling", func(t *testing.T) {
		utc := time.Now().UTC()
		local := utc.Local()

		// Should represent the same moment in time
		if utc.Unix() != local.Unix() {
			t.Error("UTC and local time should represent the same moment")
		}
	})
}

// Test ID generation patterns
func TestIDGeneration(t *testing.T) {
	t.Run("UUID-like ID Format", func(t *testing.T) {
		// Test a UUID-like format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		timestamp := time.Now().Unix()
		nano := time.Now().UnixNano() % 0x1000000000000 // Limit to 12 hex digits
		id := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			timestamp,
			1234,
			5678,
			9012,
			nano)

		parts := strings.Split(id, "-")
		if len(parts) != 5 {
			t.Errorf("UUID should have 5 parts separated by '-', got %d", len(parts))
		}

		expectedLengths := []int{8, 4, 4, 4, 12}
		for i, part := range parts {
			if len(part) != expectedLengths[i] {
				t.Errorf("Part %d should be %d characters, got %d", i, expectedLengths[i], len(part))
			}
		}
	})

	t.Run("Timestamp-based ID", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		id := fmt.Sprintf("id_%d", timestamp)

		if !strings.HasPrefix(id, "id_") {
			t.Error("Timestamp-based ID should start with 'id_'")
		}

		if len(id) < 15 { // "id_" + at least 10 digits
			t.Error("Timestamp-based ID should be at least 15 characters")
		}
	})

	t.Run("Random String Generation", func(t *testing.T) {
		// Test generating random-like strings
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length := 16

		// Simulate random string generation using time-based seed
		seed := time.Now().UnixNano()
		result := ""
		for i := 0; i < length; i++ {
			result += string(chars[(seed+int64(i))%int64(len(chars))])
		}

		if len(result) != length {
			t.Errorf("Generated string should be %d characters, got %d", length, len(result))
		}

		// Check all characters are valid
		for _, char := range result {
			if !strings.ContainsRune(chars, char) {
				t.Errorf("Invalid character in generated string: %c", char)
			}
		}
	})
}

// Test string manipulation utilities
func TestStringUtils(t *testing.T) {
	t.Run("String Sanitization", func(t *testing.T) {
		input := "  Hello, World!  \n\t"
		sanitized := strings.TrimSpace(input)

		if sanitized != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got '%s'", sanitized)
		}
	})

	t.Run("String Case Conversion", func(t *testing.T) {
		input := "Test String"
		lower := strings.ToLower(input)
		upper := strings.ToUpper(input)

		if lower != "test string" {
			t.Errorf("Expected 'test string', got '%s'", lower)
		}

		if upper != "TEST STRING" {
			t.Errorf("Expected 'TEST STRING', got '%s'", upper)
		}
	})

	t.Run("String Replacement", func(t *testing.T) {
		input := "Hello World"
		replaced := strings.ReplaceAll(input, "World", "Go")

		if replaced != "Hello Go" {
			t.Errorf("Expected 'Hello Go', got '%s'", replaced)
		}
	})

	t.Run("String Splitting", func(t *testing.T) {
		input := "apple,banana,cherry"
		parts := strings.Split(input, ",")

		if len(parts) != 3 {
			t.Errorf("Expected 3 parts, got %d", len(parts))
		}

		expected := []string{"apple", "banana", "cherry"}
		for i, part := range parts {
			if part != expected[i] {
				t.Errorf("Expected '%s', got '%s'", expected[i], part)
			}
		}
	})

	t.Run("String Joining", func(t *testing.T) {
		parts := []string{"apple", "banana", "cherry"}
		joined := strings.Join(parts, ",")

		if joined != "apple,banana,cherry" {
			t.Errorf("Expected 'apple,banana,cherry', got '%s'", joined)
		}
	})
}

// Test number formatting and validation
func TestNumberUtils(t *testing.T) {
	t.Run("Integer Formatting", func(t *testing.T) {
		num := 12345
		formatted := fmt.Sprintf("%d", num)

		if formatted != "12345" {
			t.Errorf("Expected '12345', got '%s'", formatted)
		}
	})

	t.Run("Float Formatting", func(t *testing.T) {
		num := 123.456
		formatted := fmt.Sprintf("%.2f", num)

		if formatted != "123.46" {
			t.Errorf("Expected '123.46', got '%s'", formatted)
		}
	})

	t.Run("Number Validation", func(t *testing.T) {
		validNumbers := []string{"123", "0", "-456", "78.90"}
		invalidNumbers := []string{"abc", "12.34.56", "", "12a34"}

		for _, numStr := range validNumbers {
			if !isNumeric(numStr) {
				t.Errorf("'%s' should be considered numeric", numStr)
			}
		}

		for _, numStr := range invalidNumbers {
			if isNumeric(numStr) {
				t.Errorf("'%s' should not be considered numeric", numStr)
			}
		}
	})
}

// Test array/slice utilities
func TestArrayUtils(t *testing.T) {
	t.Run("Array Contains", func(t *testing.T) {
		arr := []string{"apple", "banana", "cherry"}

		if !contains(arr, "banana") {
			t.Error("Array should contain 'banana'")
		}

		if contains(arr, "grape") {
			t.Error("Array should not contain 'grape'")
		}
	})

	t.Run("Array Unique", func(t *testing.T) {
		arr := []string{"apple", "banana", "apple", "cherry", "banana"}
		unique := removeDuplicates(arr)

		if len(unique) != 3 {
			t.Errorf("Expected 3 unique items, got %d", len(unique))
		}

		// Check that all original unique items are present
		expected := []string{"apple", "banana", "cherry"}
		for _, item := range expected {
			if !contains(unique, item) {
				t.Errorf("Unique array should contain '%s'", item)
			}
		}
	})

	t.Run("Array Filtering", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		evens := filterEvens(numbers)

		expectedEvens := []int{2, 4, 6, 8, 10}
		if len(evens) != len(expectedEvens) {
			t.Errorf("Expected %d even numbers, got %d", len(expectedEvens), len(evens))
		}

		for i, even := range evens {
			if even != expectedEvens[i] {
				t.Errorf("Expected %d, got %d", expectedEvens[i], even)
			}
		}
	})
}

// Helper functions for testing
func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	hasDigit := false
	hasDot := false

	for i, char := range s {
		if char >= '0' && char <= '9' {
			hasDigit = true
		} else if char == '.' {
			if hasDot {
				return false // Multiple dots
			}
			hasDot = true
		} else if char == '-' {
			if i != 0 {
				return false // Minus not at start
			}
		} else {
			return false // Invalid character
		}
	}

	return hasDigit
}

func contains(arr []string, item string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

func removeDuplicates(arr []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range arr {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func filterEvens(numbers []int) []int {
	var evens []int
	for _, num := range numbers {
		if num%2 == 0 {
			evens = append(evens, num)
		}
	}
	return evens
}

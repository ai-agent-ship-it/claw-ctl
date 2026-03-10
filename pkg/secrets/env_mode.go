package secrets

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnvFile reads a .env file and returns the key-value pairs.
// Lines starting with # are comments. Empty lines are skipped.
func LoadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file %s: %w", path, err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes
		value = strings.Trim(value, "\"'")

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading env file: %w", err)
	}

	return result, nil
}

// FilterSecrets returns only the keys that are in the required list.
func FilterSecrets(all map[string]string, required []string) map[string]string {
	filtered := make(map[string]string)
	for _, key := range required {
		if val, ok := all[key]; ok {
			filtered[key] = val
		}
	}
	return filtered
}

// ValidateSecrets checks that all required secrets are present and non-empty.
func ValidateSecrets(secrets map[string]string, required []string) []string {
	var missing []string
	for _, key := range required {
		if val, ok := secrets[key]; !ok || val == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

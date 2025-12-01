package pi

import (
	"os"
	"strings"
)

func GetPiId() string {
	data, err := os.ReadFile("/proc/cpuinfo")

	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Serial") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return "unknown"
}

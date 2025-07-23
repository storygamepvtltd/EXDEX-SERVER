package helper

import "strings"

func ExtractSection(raw string, label string) string {
	prefix := label + ": "
	start := strings.Index(raw, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)

	end := strings.Index(raw[start:], "}") + start + 1
	// Keep extracting until closing brace is balanced (for nested JSON)
	braces := 1
	for i := end; i < len(raw); i++ {
		if raw[i] == '{' {
			braces++
		} else if raw[i] == '}' {
			braces--
			if braces == 0 {
				end = i + 1
				break
			}
		}
	}
	return raw[start:end]
}

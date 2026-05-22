package check

import (
	"regexp"
	"strings"
)

type lockIndex struct {
	Resolved map[string]map[string]bool
}

var packageKeyVersionRE = regexp.MustCompile(`^(.+)@([^@/()]+)(?:\(.+\))?$`)

func parsePnpmLock(content string) lockIndex {
	index := lockIndex{
		Resolved: map[string]map[string]bool{},
	}

	lines := strings.Split(content, "\n")
	section := ""

	for _, raw := range lines {
		line := stripCommentRight(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := leadingSpaces(line)
		trimmed := strings.TrimSpace(line)
		key, value, ok := splitYAMLKeyValue(trimmed)
		if !ok {
			continue
		}
		key = unquote(key)
		value = strings.TrimSpace(value)

		if indent == 0 {
			section = key
			continue
		}

		if (section == "packages" || section == "snapshots") && indent == 2 {
			if name, version, ok := packageFromLockKey(key); ok {
				addVersion(index.Resolved, name, version)
			}
			continue
		}

	}

	return index
}

func packageFromLockKey(key string) (string, string, bool) {
	key = strings.TrimSpace(unquote(key))
	if key == "" {
		return "", "", false
	}

	if match := packageKeyVersionRE.FindStringSubmatch(key); len(match) == 3 {
		return match[1], normalizeVersion(match[2]), true
	}
	return "", "", false
}

func splitYAMLKeyValue(line string) (string, string, bool) {
	before, after, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(before), strings.TrimSpace(after), true
}

func leadingSpaces(value string) int {
	count := 0
	for _, r := range value {
		if r != ' ' {
			break
		}
		count++
	}
	return count
}

func stripCommentRight(line string) string {
	var quote rune
	for i, r := range line {
		switch r {
		case '\'', '"':
			if quote == 0 {
				quote = r
			} else if quote == r {
				quote = 0
			}
		case '#':
			if quote == 0 {
				return line[:i]
			}
		}
	}
	return line
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return value
	}
	if (value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"') {
		return value[1 : len(value)-1]
	}
	return value
}

func normalizeVersion(value string) string {
	value = strings.TrimSpace(unquote(value))
	value = strings.TrimPrefix(value, "link:")
	value = strings.TrimPrefix(value, "workspace:")
	value = strings.TrimPrefix(value, "npm:")
	if cut := strings.IndexByte(value, '('); cut >= 0 {
		value = value[:cut]
	}
	if cut := strings.Index(value, " @ "); cut >= 0 {
		value = value[cut+3:]
	}
	return strings.TrimSpace(value)
}

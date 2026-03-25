package classify

import "strings"

var wrappers = map[string]struct{}{
	"builtin": {},
	"command": {},
	"noglob":  {},
	"time":    {},
}

func CommandName(commandLine string) string {
	for _, token := range strings.Fields(commandLine) {
		if strings.Contains(token, "=") && !strings.HasPrefix(token, "=") && !strings.HasPrefix(token, "/") {
			parts := strings.SplitN(token, "=", 2)
			if parts[0] != "" {
				continue
			}
		}
		if _, ok := wrappers[token]; ok {
			continue
		}
		return token
	}
	return ""
}

func Category(commandLine string, categories map[string][]string) string {
	commandName := CommandName(commandLine)
	if commandName == "" {
		return "misc"
	}

	fields := strings.Fields(commandLine)
	if commandName == "go" && len(fields) > 1 {
		switch fields[len(fields)-len(fields)+1] {
		case "test":
			return "test"
		case "build":
			return "build"
		}
	}

	for _, category := range []string{"git", "test", "build", "search", "edit", "nav"} {
		for _, candidate := range categories[category] {
			if candidate == commandName {
				return category
			}
		}
	}

	return "misc"
}

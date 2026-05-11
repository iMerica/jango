package urls

import (
	"fmt"
	"regexp"
	"strings"
)

type Converter string

const (
	StringConverter Converter = "str"
	IntConverter    Converter = "int"
	SlugConverter   Converter = "slug"
	UUIDConverter   Converter = "uuid"
	PathConverter   Converter = "path"
)

type ConverterDef struct {
	Pattern   string
	Converter func(string) (string, error)
}

var Converters = map[Converter]ConverterDef{
	StringConverter: {
		Pattern:   `[^/]+`,
		Converter: stringConverter,
	},
	IntConverter: {
		Pattern:   `\d+`,
		Converter: intConverter,
	},
	SlugConverter: {
		Pattern:   `[-a-zA-Z0-9_]+`,
		Converter: slugConverter,
	},
	UUIDConverter: {
		Pattern:   `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		Converter: uuidConverter,
	},
	PathConverter: {
		Pattern:   `.+`,
		Converter: stringConverter,
	},
}

func stringConverter(v string) (string, error) {
	return v, nil
}

func intConverter(v string) (string, error) {
	if v == "" {
		return "", fmt.Errorf("empty int value")
	}
	if v[0] == '0' && len(v) > 1 {
		return "", fmt.Errorf("invalid int: leading zero")
	}
	for _, c := range v {
		if c < '0' || c > '9' {
			return "", fmt.Errorf("invalid int: %s", v)
		}
	}
	return v, nil
}

func slugConverter(v string) (string, error) {
	if v == "" {
		return "", fmt.Errorf("empty slug")
	}
	for _, c := range v {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return "", fmt.Errorf("invalid slug: %s", v)
		}
	}
	return v, nil
}

func uuidConverter(v string) (string, error) {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(v) {
		return "", fmt.Errorf("invalid UUID: %s", v)
	}
	return v, nil
}

type paramSpec struct {
	Name      string
	Converter Converter
}

type Pattern struct {
	Path        string
	Regex       *regexp.Regexp
	Handler     interface{}
	Metadata    interface{}
	Name        string
	Namespace   string
	SubPatterns []Pattern
	IsInclude   bool
	prefix      string
	paramSpecs  []paramSpec
}

func Path(route string, handler interface{}, name string) Pattern {
	regex, specs := compilePathPattern(route)
	return Pattern{
		Path:       route,
		Regex:      regex,
		Handler:    handler,
		Name:       name,
		paramSpecs: specs,
	}
}

func RePath(route string, handler interface{}, name string) Pattern {
	specs := extractNamedGroups(route)
	regex := regexp.MustCompile("^" + route + "$")
	return Pattern{
		Path:       route,
		Regex:      regex,
		Handler:    handler,
		Name:       name,
		paramSpecs: specs,
	}
}

func Include(prefix string, patterns []Pattern, namespace string) Pattern {
	return Pattern{
		Path:        prefix,
		Handler:     nil,
		Name:        "",
		Namespace:   namespace,
		SubPatterns: patterns,
		IsInclude:   true,
		prefix:      strings.TrimSuffix(prefix, "/"),
	}
}

func URLPatterns(patterns ...Pattern) []Pattern {
	return patterns
}

func compilePathPattern(pattern string) (*regexp.Regexp, []paramSpec) {
	var specs []paramSpec
	regexStr := "^"
	remaining := pattern
	idx := 0

	for idx < len(remaining) {
		if remaining[idx] == '<' {
			closeIdx := strings.Index(remaining[idx:], ">")
			if closeIdx == -1 {
				regexStr += regexp.QuoteMeta(remaining[idx:])
				break
			}
			closeIdx += idx
			spec := parseParamSpec(remaining[idx+1 : closeIdx])
			specs = append(specs, spec)
			convDef, ok := Converters[spec.Converter]
			if !ok {
				convDef = Converters[StringConverter]
			}
			regexStr += "(?P<" + spec.Name + ">" + convDef.Pattern + ")"
			idx = closeIdx + 1
		} else {
			regexStr += regexp.QuoteMeta(string(remaining[idx]))
			idx++
		}
	}

	regexStr += "$"
	return regexp.MustCompile(regexStr), specs
}

func parseParamSpec(spec string) paramSpec {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) == 2 {
		return paramSpec{
			Name:      parts[1],
			Converter: Converter(parts[0]),
		}
	}
	return paramSpec{
		Name:      parts[0],
		Converter: StringConverter,
	}
}

func extractNamedGroups(pattern string) []paramSpec {
	re := regexp.MustCompile(`\(\?P<(\w+)>`)
	matches := re.FindAllStringSubmatch(pattern, -1)
	var specs []paramSpec
	for _, m := range matches {
		specs = append(specs, paramSpec{
			Name:      m[1],
			Converter: StringConverter,
		})
	}
	return specs
}

package urls

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Match struct {
	Handler interface{}
	Params  map[string]string
	Name    string
}

type Resolver struct {
	patterns []Pattern
	names    map[string]namedEntry
}

type namedEntry struct {
	Pattern    *regexp.Regexp
	paramSpecs []paramSpec
	Namespace  string
	Prefix     string
}

func NewResolver(patterns []Pattern) *Resolver {
	r := &Resolver{
		patterns: patterns,
		names:    make(map[string]namedEntry),
	}
	r.indexPatterns(patterns, "")
	return r
}

func (r *Resolver) Resolve(path string) (*Match, error) {
	return r.resolvePatterns(r.patterns, path, "")
}

func (r *Resolver) resolvePatterns(patterns []Pattern, path string, prefix string) (*Match, error) {
	for _, pattern := range patterns {
		if pattern.IsInclude {
			match, err := r.resolveInclude(pattern, path, prefix)
			if err == nil {
				return match, nil
			}
			continue
		}

		match, err := r.matchPattern(pattern, path)
		if err == nil {
			return match, nil
		}
	}
	return nil, fmt.Errorf("no matching URL pattern for path: %s", path)
}

func (r *Resolver) resolveInclude(pattern Pattern, path string, parentPrefix string) (*Match, error) {
	matchPrefix := pattern.prefix

	if !strings.HasPrefix(path, matchPrefix+"/") && path != matchPrefix {
		return nil, fmt.Errorf("path does not match include prefix")
	}

	remaining := strings.TrimPrefix(path, matchPrefix)
	if remaining == "" || remaining == "/" {
		remaining = "/"
	} else if !strings.HasPrefix(remaining, "/") {
		remaining = "/" + remaining
	}

	fullPrefix := matchPrefix
	if parentPrefix != "" {
		fullPrefix = strings.TrimSuffix(parentPrefix, "/") + "/" + strings.TrimPrefix(matchPrefix, "/")
	}

	match, err := r.resolvePatterns(pattern.SubPatterns, remaining, fullPrefix)
	if err != nil {
		return nil, err
	}

	ns := pattern.Namespace
	if ns != "" && match.Name != "" {
		match.Name = ns + ":" + match.Name
	}

	return match, nil
}

func (r *Resolver) matchPattern(pattern Pattern, path string) (*Match, error) {
	if pattern.Regex == nil {
		return nil, fmt.Errorf("pattern has no regex")
	}

	if !pattern.Regex.MatchString(path) {
		return nil, fmt.Errorf("pattern does not match path: %s", path)
	}

	params := extractParams(pattern.Regex, path, pattern.paramSpecs)

	if len(params) > 0 {
		for _, spec := range pattern.paramSpecs {
			val, ok := params[spec.Name]
			if !ok {
				continue
			}
			convDef, exists := Converters[spec.Converter]
			if !exists {
				convDef = Converters[StringConverter]
			}
			converted, err := convDef.Converter(val)
			if err != nil {
				return nil, fmt.Errorf("converter error for param %s: %v", spec.Name, err)
			}
			params[spec.Name] = converted
		}
	}

	return &Match{
		Handler: pattern.Handler,
		Params:  params,
		Name:    pattern.Name,
	}, nil
}

func extractParams(re *regexp.Regexp, path string, specs []paramSpec) map[string]string {
	match := re.FindStringSubmatch(path)
	if match == nil {
		return nil
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		if i < len(match) {
			result[name] = match[i]
		}
	}
	return result
}

func (r *Resolver) indexPatterns(patterns []Pattern, prefix string) {
	for i := range patterns {
		pattern := &patterns[i]
		if pattern.IsInclude {
			newPrefix := pattern.prefix
			if prefix != "" {
				newPrefix = strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(newPrefix, "/")
			}
			r.indexPatterns(pattern.SubPatterns, newPrefix)
			continue
		}
		if pattern.Name != "" {
			r.names[pattern.Name] = namedEntry{
				Pattern:    pattern.Regex,
				paramSpecs: pattern.paramSpecs,
				Namespace:  "",
				Prefix:     prefix,
			}
		}
	}
}

func (r *Resolver) Reverse(name string, kwargs map[string]string) (string, error) {
	if entry, ok := r.names[name]; ok {
		return r.buildReverse(entry, kwargs)
	}

	if strings.Contains(name, ":") {
		parts := strings.SplitN(name, ":", 2)
		localName := parts[1]
		if entry, ok := r.names[localName]; ok {
			return r.buildReverse(entry, kwargs)
		}
	}

	return "", fmt.Errorf("no URL pattern named %q", name)
}

func (r *Resolver) buildReverse(entry namedEntry, kwargs map[string]string) (string, error) {
	path := entry.Pattern.String()
	path = strings.TrimPrefix(path, "^")
	path = strings.TrimSuffix(path, "$")

	re := regexp.MustCompile(`\(\?P<(\w+)>[^)]+\)`)
	matches := re.FindAllStringSubmatchIndex(path, -1)

	for i := len(matches) - 1; i >= 0; i-- {
		loc := matches[i]
		paramExpr := path[loc[0]:loc[1]]
		nameRe := regexp.MustCompile(`\(\?P<(\w+)>`)
		nameMatch := nameRe.FindStringSubmatch(paramExpr)
		if len(nameMatch) < 2 {
			continue
		}
		paramName := nameMatch[1]
		val, ok := kwargs[paramName]
		if !ok {
			return "", fmt.Errorf("missing parameter %q for reverse URL", paramName)
		}
		path = path[:loc[0]] + val + path[loc[1]:]
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if entry.Prefix != "" {
		path = strings.TrimSuffix(entry.Prefix, "/") + path
	}

	return path, nil
}

func (r *Resolver) Patterns() []Pattern {
	return r.patterns
}

func (r *Resolver) Register(pattern Pattern) {
	r.patterns = append(r.patterns, pattern)
	if pattern.Name != "" && !pattern.IsInclude {
		r.names[pattern.Name] = namedEntry{
			Pattern:    pattern.Regex,
			paramSpecs:  pattern.paramSpecs,
			Namespace:  "",
		}
	}
}

func Handle(pattern Pattern, w http.ResponseWriter, r *http.Request) {
	if handler, ok := pattern.Handler.(http.Handler); ok {
		handler.ServeHTTP(w, r)
		return
	}
	if fn, ok := pattern.Handler.(http.HandlerFunc); ok {
		fn(w, r)
		return
	}
	http.NotFound(w, r)
}
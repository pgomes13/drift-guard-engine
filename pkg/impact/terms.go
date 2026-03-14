package impact

import (
	"strings"

	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

// ExtractTerms returns search terms to grep for when scanning source files for
// references to the given breaking change.
func ExtractTerms(c schema.Change) []string {
	var terms []string
	seen := map[string]bool{}
	add := func(t string) {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			seen[t] = true
			terms = append(terms, t)
		}
	}

	ct := string(c.Type)
	switch {
	case strings.HasPrefix(ct, "grpc_"):
		// Location is "ServiceName", "ServiceName.RPCName", or "MessageName.fieldName"
		parts := strings.SplitN(c.Location, ".", 2)
		add(parts[0])
		if len(parts) == 2 {
			add(parts[1])
		}

	case strings.HasPrefix(ct, "gql_"):
		// Location is "TypeName" or "TypeName.fieldName"
		parts := strings.SplitN(c.Location, ".", 2)
		add(parts[0])
		if len(parts) == 2 {
			add(parts[1])
		}

	default:
		// OpenAPI: use the path and any field/param name from Location
		if c.Path != "" {
			add(simplifyOpenAPIPath(c.Path))
		}
		if c.Location != "" {
			// Location examples: "param.query.limit", "request.body.email", "response.200.id"
			// Extract the last segment as the field/param name.
			parts := strings.Split(c.Location, ".")
			last := parts[len(parts)-1]
			if !isNumeric(last) && last != "type" && last != "required" {
				add(last)
			}
		}
	}

	return terms
}

// simplifyOpenAPIPath strips path-parameter segments from an OpenAPI path and
// returns a stable prefix suitable for grepping.
// "/users/{id}/posts/{postId}" → "/users/"
func simplifyOpenAPIPath(path string) string {
	parts := strings.Split(path, "/")
	var out []string
	for _, p := range parts {
		if strings.HasPrefix(p, "{") {
			break
		}
		out = append(out, p)
	}
	result := strings.Join(out, "/")
	if result == "" || result == "/" {
		return "/"
	}
	if !strings.HasSuffix(result, "/") {
		result += "/"
	}
	return result
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

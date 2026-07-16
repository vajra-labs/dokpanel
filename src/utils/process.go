package utils

import (
	"crypto/rand"
	"encoding/json/v2"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Schema holds the context needed for domain generation.
type Schema struct {
	ServerIP    string
	ProjectName string
}

// ProcessedTemplate is the output after processing a CompleteTemplate.
type ProcessedTemplate struct {
	Domains []DomainConfig
	Envs    []string
	Mounts  []MountConfig
}

// ProcessTemplate is the main entry point.
func ProcessTemplate(t *CompleteTemplate, schema Schema) *ProcessedTemplate {
	variables := processVariables(t, schema)
	return &ProcessedTemplate{
		Domains: processDomains(t, variables, schema),
		Envs:    processEnvVars(t, variables, schema),
		Mounts:  processMounts(t, variables, schema),
	}
}

// processVariables resolves all template variables.
func processVariables(t *CompleteTemplate, schema Schema) map[string]string {
	variables := make(map[string]string, len(t.Variables))
	// First pass: resolve generator placeholders
	for key, value := range t.Variables {
		switch {
		case value == "${domain}":
			variables[key] = GenerateRandomDomain(schema.ServerIP, schema.ProjectName)
		case value == "${base64}":
			variables[key] = GenerateBase64Str(32)
		case strings.HasPrefix(value, "${base64:"):
			length := extractInt(value, `\$\{base64:(\d+)\}`, 32)
			variables[key] = GenerateBase64Str(length)
		case value == "${password}":
			variables[key] = GeneratePassword(16)
		case strings.HasPrefix(value, "${password:"):
			length := extractInt(value, `\$\{password:(\d+)\}`, 16)
			variables[key] = GeneratePassword(length)
		case value == "${hash}":
			variables[key] = GenerateHash(8)
		case strings.HasPrefix(value, "${hash:"):
			length := extractInt(value, `\$\{hash:(\d+)\}`, 8)
			variables[key] = GenerateHash(length)
		default:
			variables[key] = value
		}
	}

	// Second pass: resolve cross-variable references
	for key, value := range variables {
		variables[key] = processValue(value, variables, schema)
	}

	return variables
}

// placeholderRe matches ${...} placeholders inside strings.
var placeholderRe = regexp.MustCompile(`\$\{([^}]+)\}`)

// processValue replaces all ${...} in a string with resolved values.
func processValue(value string, variables map[string]string, schema Schema) string {
	return placeholderRe.ReplaceAllStringFunc(value, func(match string) string {
		inner := match[2 : len(match)-1] // strip ${ and }
		return resolvePlaceholder(inner, variables, schema)
	})
}

// resolvePlaceholder maps a placeholder name to its generated/resolved value.
func resolvePlaceholder(varName string, variables map[string]string, schema Schema) string {
	switch {
	case varName == "domain":
		return GenerateRandomDomain(schema.ServerIP, schema.ProjectName)

	case varName == "base64":
		return GenerateBase64Str(32)
	case strings.HasPrefix(varName, "base64:"):
		length, _ := strconv.Atoi(strings.TrimPrefix(varName, "base64:"))
		return GenerateBase64Str(length)

	case varName == "password":
		return GeneratePassword(16)
	case strings.HasPrefix(varName, "password:"):
		length, _ := strconv.Atoi(strings.TrimPrefix(varName, "password:"))
		return GeneratePassword(length)

	case varName == "hash":
		return GenerateHash(8)
	case strings.HasPrefix(varName, "hash:"):
		length, _ := strconv.Atoi(strings.TrimPrefix(varName, "hash:"))
		return GenerateHash(length)

	case varName == "uuid":
		return uuid.New().String()

	case varName == "timestamp", varName == "timestampms":
		return strconv.FormatInt(time.Now().UnixMilli(), 10)

	case varName == "timestamps":
		return strconv.FormatInt(time.Now().Unix(), 10)

	case strings.HasPrefix(varName, "timestampms:"):
		t, err := time.Parse(time.RFC3339, strings.TrimPrefix(varName, "timestampms:"))
		if err != nil {
			return "${" + varName + "}"
		}
		return strconv.FormatInt(t.UnixMilli(), 10)

	case strings.HasPrefix(varName, "timestamps:"):
		t, err := time.Parse(time.RFC3339, strings.TrimPrefix(varName, "timestamps:"))
		if err != nil {
			return "${" + varName + "}"
		}
		return strconv.FormatInt(t.Unix(), 10)

	case varName == "randomPort":
		n, _ := rand.Int(rand.Reader, big.NewInt(65535))
		return n.String()

	case varName == "jwt":
		return GenerateJWT("", nil, 0)

	case strings.HasPrefix(varName, "jwt:"):
		return resolveJWT(varName, variables)

	case varName == "username":
		return generateRandomUsername()

	case varName == "email":
		return generateRandomEmail()

	default:
		if v, ok := variables[varName]; ok {
			return v
		}
		return "${" + varName + "}" // unresolved — keep as-is
	}
}

// resolveJWT handles jwt:length and jwt:secret:payload variants.
func resolveJWT(varName string, variables map[string]string) string {
	parts := strings.SplitN(strings.TrimPrefix(varName, "jwt:"), ":", 2)
	// jwt:64 — random hex of given length
	if len(parts) == 1 {
		if length, err := strconv.Atoi(parts[0]); err == nil {
			return GenerateJWT("", nil, length)
		}
	}
	// jwt:SECRET_VAR or jwt:SECRET_VAR:{"role":"admin"}
	secret := ""
	if len(parts) >= 1 {
		secretKey := parts[0]
		if v, ok := variables[secretKey]; ok {
			secret = v
		} else {
			secret = secretKey
		}
	}
	var payloadMap map[string]any
	if len(parts) >= 2 {
		payloadStr := parts[1]
		if v, ok := variables[payloadStr]; ok {
			payloadStr = v
		}
		if strings.HasPrefix(strings.TrimSpace(payloadStr), "{") {
			_ = json.Unmarshal([]byte(payloadStr), &payloadMap)
		}
	}
	return GenerateJWT(secret, payloadMap, 0)
}

// processDomains resolves domain host placeholders.
func processDomains(
	t *CompleteTemplate,
	variables map[string]string,
	schema Schema,
) []DomainConfig {
	if len(t.Config.Domains) == 0 {
		return []DomainConfig{}
	}
	out := make([]DomainConfig, 0, len(t.Config.Domains))
	for _, d := range t.Config.Domains {
		if d.ServiceName == "" {
			continue
		}
		host := d.Host
		if host == "" {
			host = GenerateRandomDomain(schema.ServerIP, schema.ProjectName)
		} else {
			host = processValue(host, variables, schema)
		}
		out = append(out, DomainConfig{
			ServiceName: d.ServiceName,
			Port:        d.Port,
			Path:        d.Path,
			Host:        host,
		})
	}
	return out
}

// processEnvVars resolves env var values with placeholder substitution.
func processEnvVars(t *CompleteTemplate, variables map[string]string, schema Schema) []string {
	if len(t.Config.Env) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(t.Config.Env))
	for key, rawVal := range t.Config.Env {
		var valStr string
		switch v := rawVal.(type) {
		case string:
			valStr = processValue(v, variables, schema)
		case bool:
			valStr = strconv.FormatBool(v)
		case float64:
			valStr = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			valStr = fmt.Sprintf("%v", v)
		}
		out = append(out, fmt.Sprintf("%s=%s", key, valStr))
	}
	return out
}

// processMounts resolves mount filePath and content placeholders.
func processMounts(t *CompleteTemplate, variables map[string]string, schema Schema) []MountConfig {
	if len(t.Config.Mounts) == 0 {
		return []MountConfig{}
	}

	out := make([]MountConfig, 0, len(t.Config.Mounts))
	for _, m := range t.Config.Mounts {
		if m.FilePath == "" && m.Content == "" {
			continue
		}
		out = append(out, MountConfig{
			FilePath: processValue(m.FilePath, variables, schema),
			Content:  processValue(m.Content, variables, schema),
		})
	}
	return out
}

// extractInt extracts a number from a string using regex, with a default fallback.
func extractInt(s, pattern string, defaultVal int) int {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if len(match) < 2 {
		return defaultVal
	}
	n, err := strconv.Atoi(match[1])
	if err != nil {
		return defaultVal
	}
	return n
}

// generateRandomUsername returns a simple random username (replaces faker).
func generateRandomUsername() string {
	adjectives := []string{"fast", "cool", "happy", "bright", "calm", "bold", "swift", "keen"}
	nouns := []string{"tiger", "eagle", "wolf", "panda", "hawk", "lion", "bear", "fox"}
	n1, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	n2, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	return fmt.Sprintf("%s_%s_%s", adjectives[n1.Int64()], nouns[n2.Int64()], GenerateHash(4))
}

// generateRandomEmail returns a simple random email.
func generateRandomEmail() string {
	domains := []string{"example.com", "mail.io", "test.dev"}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(domains))))
	return fmt.Sprintf("%s@%s", generateRandomUsername(), domains[n.Int64()])
}

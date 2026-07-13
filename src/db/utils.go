package db

import (
	"fmt"
	"math/rand"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz123456789"

var verbs = []string{
	"compress", "connect", "copy", "generate", "parse",
	"calculate", "index", "navigate", "override", "restart",
}

var adjectives = []string{
	"virtual", "wireless", "primary", "dynamic", "auxiliary",
	"solid", "mobile", "neural", "digital", "open",
}

var nouns = []string{
	"system", "driver", "protocol", "interface", "firewall",
	"sensor", "network", "array", "application", "monitor",
}

func randomFrom(list []string) string {
	return list[rand.Intn(len(list))]
}

func generatePassword(length int) string {
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}
	return b.String()
}

// GenAppName generates:
// type-verb-adjective-noun-abcdef
func GenAppName(appType string) string {
	verb := strings.ReplaceAll(randomFrom(verbs), " ", "-")
	adjective := strings.ReplaceAll(randomFrom(adjectives), " ", "-")
	noun := strings.ReplaceAll(randomFrom(nouns), " ", "-")
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		appType,
		verb,
		adjective,
		noun,
		generatePassword(6),
	)
}

// CleanAppName trims spaces, replaces spaces with '-' and converts to lowercase.
func CleanAppName(appName string) string {
	appName = strings.TrimSpace(appName)
	appName = strings.ReplaceAll(appName, " ", "-")
	return strings.ToLower(appName)
}

// BuildAppName follows same logic as TypeScript.
func BuildAppName(appType string, baseAppName string) string {
	if strings.TrimSpace(baseAppName) != "" {
		return fmt.Sprintf("%s-%s", CleanAppName(baseAppName), generatePassword(6))
	}
	return GenAppName(appType)
}

package db

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"goploy/src/utils"
)

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
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(list))))
	return list[n.Int64()]
}

// GenAppName generates: type-verb-adjective-noun-xxxxxx
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
		utils.GenerateHash(6),
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
		return fmt.Sprintf("%s-%s", CleanAppName(baseAppName), utils.GeneratePassword(6))
	}
	return GenAppName(appType)
}

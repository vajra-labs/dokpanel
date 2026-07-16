// Post-processes atlas-generated goose migration files.
// Wraps CREATE TRIGGER statements with goose StatementBegin/End annotations
// since atlas does not add them automatically in goose format.
//
// Usage:
//
//	go run ./sqldb/script.go               # auto-detects latest migration
//	go run ./sqldb/script.go <file>        # patch a specific file
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const migrateDir = "sqldb/migrate"

// Matches:
//
// -- create trigger "..."
// CREATE TRIGGER ... END;
var triggerRe = regexp.MustCompile(`(?is)(-- create trigger[^\n]*\n)(CREATE TRIGGER\b[\s\S]+?END;)`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	path, err := migrationPath()
	if err != nil {
		return err
	}

	if err := patchFile(path); err != nil {
		return err
	}

	fmt.Printf("✓ patched: %s\n", path)
	return nil
}

func migrationPath() (string, error) {
	if len(os.Args) >= 2 {
		return os.Args[1], nil
	}

	return latestMigration(migrateDir)
}

func patchFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	patched := wrapTriggers(string(data))

	if err := os.WriteFile(path, []byte(patched), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// latestMigration returns the newest migration file.
//
// Atlas migration filenames are timestamp-prefixed, so the
// lexicographically largest filename is the latest migration.
func latestMigration(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var latest string

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, e.Name())

		if latest == "" || path > latest {
			latest = path
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no .sql files found in %s", dir)
	}

	return latest, nil
}

// wrapTriggers wraps atlas-generated CREATE TRIGGER statements with
// goose StatementBegin/StatementEnd directives.
//
// Already wrapped triggers are left unchanged.
func wrapTriggers(content string) string {
	return triggerRe.ReplaceAllString(
		content,
		"$1-- +goose StatementBegin\n$2\n-- +goose StatementEnd",
	)
}

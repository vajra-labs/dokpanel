package provider

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
)

// GetCreateComposeFileCommand returns a bash script to write compose file contents.
func GetCreateComposeFileCommand(
	composePath string,
	appName string,
	composeFile string,
) string {
	outputPath := filepath.Join(composePath, appName, "code")
	filePath := filepath.Join(outputPath, "docker-compose.yml")
	encoded := base64.StdEncoding.EncodeToString([]byte(composeFile))

	return fmt.Sprintf(`rm -rf %q;
mkdir -p %q;
echo %q | base64 -d > %q;
echo "File 'docker-compose.yml' created: ✅";`, outputPath, outputPath, encoded, filePath)
}

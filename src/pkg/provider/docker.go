package provider

import (
	"fmt"
)

// BuildRemoteDocker returns a bash script to authenticate and pull a Docker image.
func BuildRemoteDocker(
	registryUrl, dockerImage, username, password string,
) string {
	command := fmt.Sprintf("echo %q;\n", "Pulling "+dockerImage)
	if username != "" && password != "" {
		command += fmt.Sprintf(`if ! docker login %s -u %q -p %q 2>&1; then
	echo "❌ Login failed";
	exit 1;
fi
`, registryUrl, username, password)
	}

	command += fmt.Sprintf(`docker pull %q 2>&1 || { 
  echo "❌ Pulling image failed";
  exit 1;
}
echo "✅ Pulling image completed.";`, dockerImage)

	return command
}

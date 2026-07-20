package server

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"strings"

	"goploy/src/pkg/shellx"
)

type ToolStatus struct {
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

type ServerValidationResult struct {
	Docker                   ToolStatus `json:"docker"`
	Rclone                   ToolStatus `json:"rclone"`
	Nixpacks                 ToolStatus `json:"nixpacks"`
	Buildpacks               ToolStatus `json:"buildpacks"`
	Railpack                 ToolStatus `json:"railpack"`
	IsGoployNetworkInstalled bool       `json:"isGoployNetworkInstalled"`
	IsSwarmInstalled         bool       `json:"isSwarmInstalled"`
	IsMainDirectoryInstalled bool       `json:"isMainDirectoryInstalled"`
	PrivilegeMode            string     `json:"privilegeMode"`
	DockerGroupMember        bool       `json:"dockerGroupMember"`
}

const validateDocker = `
  if command_exists docker; then
     echo "$(docker --version | awk '{print $3}' | sed 's/,//') true"
  else
    echo "0.0.0 false"
  fi
`

const validateRClone = `
  if command_exists rclone; then
    echo "$(rclone --version | head -n 1 | awk '{print $2}' | sed 's/^v//') true"
  else
    echo "0.0.0 false"
  fi
`

const validateSwarm = `
  if docker info --format '{{.Swarm.LocalNodeState}}' | grep -q 'active'; then
    echo true
  else
    echo false
  fi
`

const validateNixpacks = `
  if command_exists nixpacks; then
	version=$(nixpacks --version | awk '{print $2}')
    if [ -n "$version" ]; then
      echo "$version true"
    else
      echo "0.0.0 false"
    fi
  else
    echo "0.0.0 false"
  fi
`

const validateRailpack = `
  if command_exists railpack; then
    version=$(railpack --version | awk '{print $3}')
    if [ -n "$version" ]; then
      echo "$version true"
    else
      echo "0.0.0 false"
    fi
  else
    echo "0.0.0 false"
  fi
`

const validateBuildpacks = `
  if command_exists pack; then
    version=$(pack --version | awk '{print $1}')
    if [ -n "$version" ]; then
      echo "$version true"
    else
      echo "0.0.0 false"
    fi
  else
    echo "0.0.0 false"
  fi
`

const validateMainDirectory = `
  if [ -d "/etc/goploy" ]; then
	echo true
  else
	echo false
  fi
`

const validateGoployNetwork = `
  if docker network ls | grep -q 'goploy-network'; then
	echo true
  else
	echo false
  fi
`

const validateSudoAccess = `
  if [ "$(id -u)" -eq 0 ]; then
    echo "root true"
  elif sudo -n true 2>/dev/null; then
    echo "sudo true"
  else
    echo "none false"
  fi
`

const validateDockerGroup = `
  if groups | grep -qw docker; then
    echo true
  else
    echo false
  fi
`

// ValidateServer queries and validates the server state, checking for required tools and access.
func ValidateServer(
	ctx context.Context,
	pool *shellx.SSHPool,
	serverId int64,
) (*ServerValidationResult, error) {
	// Construct the combined bash validation command
	bashCommand := fmt.Sprintf(`
command_exists() {
  command -v "$@" > /dev/null 2>&1
}

dockerVersionEnabled=$( %s )
rcloneVersionEnabled=$( %s )
nixpacksVersionEnabled=$( %s )
buildpacksVersionEnabled=$( %s )
railpackVersionEnabled=$( %s )

dockerVersion=$(echo $dockerVersionEnabled | awk '{print $1}')
dockerEnabled=$(echo $dockerVersionEnabled | awk '{print $2}')

rcloneVersion=$(echo $rcloneVersionEnabled | awk '{print $1}')
rcloneEnabled=$(echo $rcloneVersionEnabled | awk '{print $2}')

nixpacksVersion=$(echo $nixpacksVersionEnabled | awk '{print $1}')
nixpacksEnabled=$(echo $nixpacksVersionEnabled | awk '{print $2}')

railpackVersion=$(echo $railpackVersionEnabled | awk '{print $1}')
railpackEnabled=$(echo $railpackVersionEnabled | awk '{print $2}')

buildpacksVersion=$(echo $buildpacksVersionEnabled | awk '{print $1}')
buildpacksEnabled=$(echo $buildpacksVersionEnabled | awk '{print $2}')

isGoployNetworkInstalled=$( %s )
isSwarmInstalled=$( %s )
isMainDirectoryInstalled=$( %s )

sudoAccessResult=$( %s )
privilegeMode=$(echo $sudoAccessResult | awk '{print $1}')
isDockerGroupMember=$( %s )

echo "{\"docker\": {\"version\": \"$dockerVersion\", \"enabled\": $dockerEnabled}, \"rclone\": {\"version\": \"$rcloneVersion\", \"enabled\": $rcloneEnabled}, \"nixpacks\": {\"version\": \"$nixpacksVersion\", \"enabled\": $nixpacksEnabled}, \"buildpacks\": {\"version\": \"$buildpacksVersion\", \"enabled\": $buildpacksEnabled}, \"railpack\": {\"version\": \"$railpackVersion\", \"enabled\": $railpackEnabled}, \"isGoployNetworkInstalled\": $isGoployNetworkInstalled, \"isSwarmInstalled\": $isSwarmInstalled, \"isMainDirectoryInstalled\": $isMainDirectoryInstalled, \"privilegeMode\": \"$privilegeMode\", \"dockerGroupMember\": $isDockerGroupMember}"
`,
		validateDocker,
		validateRClone,
		validateNixpacks,
		validateBuildpacks,
		validateRailpack,
		validateGoployNetwork,
		validateSwarm,
		validateMainDirectory,
		validateSudoAccess,
		validateDockerGroup,
	)
	// Execute via SSHPool
	resChan := pool.Exec(ctx, serverId, bashCommand, nil)
	res := <-resChan
	if res.Err != nil {
		return nil, res.Err
	}
	stdout := strings.TrimSpace(res.Stdout)
	var result ServerValidationResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return nil, fmt.Errorf("failed to parse server validation output %q: %w", stdout, err)
	}
	return &result, nil
}

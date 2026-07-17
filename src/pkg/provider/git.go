package provider

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"goploy/src/db/repos"
)

var sshPathRegex = regexp.MustCompile(
	`(?i)^\s*(?:([a-z]+)://)?(?:([a-z_][a-z0-9_-]*)@)?([^\s/?#:]+)(?::([0-9]{1,5}))?(?:[/:]([^\s/?#:]+))?[/:]([^\s?#:]+)`,
)

type CloneGitParams struct {
	AppName            string
	CustomGitUrl       string
	CustomGitBranch    string
	CustomGitSSHKeyId  *int64 // SSH Key ID in the database
	EnableSubmodules   bool
	Type               string // "application" | "compose"
	BasePaths          string // base path for application/compose
	SSHPath            string // path to ssh configs (for known_hosts)
	OutputPathOverride string
}

// CloneGitRepository returns a bash script to clone a custom Git repository,
// querying the database for SSH keys and updating their last used timestamp.
func CloneGitRepository(
	ctx context.Context,
	queries *repos.Queries,
	params CloneGitParams,
) (string, error) {
	command := "set -e;\n"
	if params.CustomGitUrl == "" || params.CustomGitBranch == "" {
		return "", fmt.Errorf("Error: ❌ Repository URL or Branch is missing")
	}

	outputPath := params.OutputPathOverride
	if outputPath == "" {
		outputPath = filepath.Join(params.BasePaths, params.AppName, "code")
	}

	isSsh := !strings.HasPrefix(params.CustomGitUrl, "http://") &&
		!strings.HasPrefix(params.CustomGitUrl, "https://")

	if isSsh {
		if params.CustomGitSSHKeyId == nil {
			return "", fmt.Errorf("Error: ❌ SSH Key is required for SSH clone")
		}

		// Parse SSH path once to avoid redundant regex operations and fail early on errors
		details, err := sanitizeRepoPathSSH(params.CustomGitUrl)
		if err != nil {
			return "", fmt.Errorf("Error: ❌ Invalid SSH URL: %w", err)
		}

		// 1. Fetch SSH Key from database (matches findSSHKeyById)
		sshKey, err := queries.GetSSHKeyByID(ctx, *params.CustomGitSSHKeyId)
		if err != nil {
			return "", fmt.Errorf("Error: ❌ SSH Key not found: %w", err)
		}

		// 2. Update SSH Key lastUsedAt (matches updateSSHKeyById)
		now := time.Now().Unix()
		_, err = queries.UpdateSSHKey(ctx, repos.UpdateSSHKeyParams{
			ID:          sshKey.ID,
			Name:        sshKey.Name,
			Description: sshKey.Description,
			PrivateKey:  sshKey.PrivateKey,
			PublicKey:   sshKey.PublicKey,
			LastUsedAt:  &now,
		})
		if err != nil {
			// Non-blocking but log it or return error
			return "", fmt.Errorf(
				"failed to update SSH Key last used timestamp: %w",
				err,
			)
		}

		// Write temporary key to a unique path per application to prevent race conditions in parallel builds
		keyPath := filepath.Join(
			"/tmp",
			fmt.Sprintf("id_rsa_%s", params.AppName),
		)
		command += fmt.Sprintf("trap 'rm -f %q' EXIT;\n", keyPath)
		command += fmt.Sprintf(`cat << 'EOF' > %[1]q
%[2]s
EOF
chmod 600 %[1]q;
`, keyPath, sshKey.PrivateKey)

		// Set GIT_SSH_COMMAND to accept new host keys automatically using unique key path
		portStr := fmt.Sprintf(" -p %d", details.Port)
		knownHostsPath := filepath.Join(params.SSHPath, "known_hosts")
		command += fmt.Sprintf(
			`mkdir -p "%[1]s";
export GIT_SSH_COMMAND="ssh -i %[2]q%[3]s -o ConnectTimeout=15 -o UserKnownHostsFile=%[4]q -o StrictHostKeyChecking=accept-new";
`,
			params.SSHPath,
			keyPath,
			portStr,
			knownHostsPath,
		)

		// Run ssh-keyscan to populate known_hosts (best-effort)
		command += fmt.Sprintf(
			"ssh-keyscan -p %d %s >> %q || true;\n",
			details.Port,
			details.Domain,
			knownHostsPath,
		)
	}

	submodulesFlag := ""
	submodulesUpdateCmd := ""
	if params.EnableSubmodules {
		submodulesFlag = "--recurse-submodules "
		submodulesUpdateCmd = "git submodule update --init --recursive;"
	}

	command += fmt.Sprintf(
		`if [ -d "%[1]s/.git" ]; then
	echo "🔄 Updating existing repository in %[1]s...";
	cd "%[1]s";
	git remote set-url origin %[2]s;
	if git fetch --prune --depth 1 origin %[3]s; then
		git reset --hard FETCH_HEAD;
		%[4]s
	else
		echo "⚠️ Fetch failed, performing clean clone...";
		cd /;
		rm -rf "%[1]s";
		mkdir -p "%[1]s";
		git clone --branch %[3]s --depth 1 %[5]s--progress %[2]s "%[1]s";
	fi
else
	echo "📥 Cloning Custom Git Repo %[2]s to %[1]s...";
	rm -rf "%[1]s";
	mkdir -p "%[1]s";
	if ! git clone --branch %[3]s --depth 1 %[5]s--progress %[2]s "%[1]s"; then
		echo "❌ [ERROR] Failed to clone repository";
		exit 1;
	fi
fi
`,
		outputPath,
		params.CustomGitUrl,
		params.CustomGitBranch,
		submodulesUpdateCmd,
		submodulesFlag,
	)

	return command, nil
}

type sshPathDetails struct {
	User     string
	Domain   string
	Port     int
	Owner    string
	Repo     string
	RepoPath string
}

// sanitizeRepoPathSSH parses and sanitizes an SSH repository path.
func sanitizeRepoPathSSH(input string) (*sshPathDetails, error) {
	matches := sshPathRegex.FindStringSubmatch(input)
	if len(matches) < 7 {
		return nil, fmt.Errorf("malformatted SSH path: %s", input)
	}

	user := matches[2]
	if user == "" {
		user = "git"
	}

	domain := matches[3]

	port := 22
	if matches[4] != "" {
		p, err := strconv.Atoi(matches[4])
		if err == nil {
			port = p
		}
	}

	owner := matches[5]

	repo := matches[6]
	repo = strings.TrimSpace(repo)
	repo = strings.TrimSuffix(repo, "/")
	repo = strings.TrimSuffix(repo, ".git")

	repoPath := fmt.Sprintf("ssh://%s@%s:%d/%s", user, domain, port, owner)
	if owner != "" {
		repoPath += "/"
	}
	repoPath += repo + ".git"

	return &sshPathDetails{
		User:     user,
		Domain:   domain,
		Port:     port,
		Owner:    owner,
		Repo:     repo,
		RepoPath: repoPath,
	}, nil
}

type GitCommitInfo struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
}

// GetGitCommitInfoCommand returns the git command to retrieve the last commit details.
func GetGitCommitInfoCommand(
	basePath string,
	appName string,
	gitType string,
) string {
	outputPath := filepath.Join(basePath, appName, "code")
	return fmt.Sprintf(
		`git -C %q log -1 --pretty=format:"%%H---DELIMITER---%%B"`,
		outputPath,
	)
}

// ParseGitCommitInfo parses the git log output into a GitCommitInfo struct.
func ParseGitCommitInfo(output string) *GitCommitInfo {
	parts := strings.Split(output, "---DELIMITER---")
	if len(parts) != 2 {
		return nil
	}
	return &GitCommitInfo{
		Hash:    strings.TrimSpace(parts[0]),
		Message: strings.TrimSpace(parts[1]),
	}
}

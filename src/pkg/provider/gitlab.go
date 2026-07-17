package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type CloneGitlabParams struct {
	AppName            string
	GitlabUrl          string // e.g. "https://gitlab.com"
	Owner              string
	Repository         string
	Branch             string
	Token              string
	EnableSubmodules   bool
	BasePaths          string
	OutputPathOverride string
}

// CloneGitlabRepository returns a bash script to clone a GitLab repository.
func CloneGitlabRepository(params CloneGitlabParams) string {
	command := "set -e;\n"

	outputPath := params.OutputPathOverride
	if outputPath == "" {
		outputPath = filepath.Join(params.BasePaths, params.AppName, "code")
	}

	gitlabHost := params.GitlabUrl
	var scheme string
	if len(gitlabHost) > 7 && gitlabHost[:7] == "http://" {
		scheme = "http://"
		gitlabHost = gitlabHost[7:]
	} else if len(gitlabHost) > 8 && gitlabHost[:8] == "https://" {
		scheme = "https://"
		gitlabHost = gitlabHost[8:]
	} else {
		scheme = "https://"
	}

	repoClone := fmt.Sprintf(
		"%s/%s/%s.git",
		gitlabHost,
		params.Owner,
		params.Repository,
	)
	cloneURL := fmt.Sprintf("%soauth2:%s@%s", scheme, params.Token, repoClone)

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
	echo "📥 Cloning GitLab Repo %[6]s to %[1]s...";
	rm -rf "%[1]s";
	mkdir -p "%[1]s";
	if ! git clone --branch %[3]s --depth 1 %[5]s--progress %[2]s "%[1]s"; then
		echo "❌ [ERROR] Failed to clone repository";
		exit 1;
	fi
fi
`,
		outputPath,
		cloneURL,
		params.Branch,
		submodulesUpdateCmd,
		submodulesFlag,
		repoClone,
	)

	return command
}

type GitlabRepo struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	Namespace         struct {
		Path string `json:"path"`
		Kind string `json:"kind"`
	} `json:"namespace"`
}

type GitlabBranch struct {
	Name string `json:"name"`
}

// GetGitlabRepositories lists all GitLab projects accessible to the user.
func GetGitlabRepositories(
	ctx context.Context,
	accessToken string,
	gitlabURL string,
) ([]GitlabRepo, error) {
	baseUrl := strings.TrimSuffix(gitlabURL, "/")
	url := fmt.Sprintf(
		"%s/api/v4/projects?membership=true&simple=true&per_page=100",
		baseUrl,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"gitlab api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var repos []GitlabRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// GetGitlabBranches lists branches of a GitLab project by project ID.
func GetGitlabBranches(
	ctx context.Context,
	accessToken string,
	gitlabURL string,
	projectID int64,
) ([]GitlabBranch, error) {
	baseUrl := strings.TrimSuffix(gitlabURL, "/")
	url := fmt.Sprintf(
		"%s/api/v4/projects/%d/repository/branches?per_page=100",
		baseUrl,
		projectID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"gitlab api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var branches []GitlabBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

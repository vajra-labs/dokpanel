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

type CloneGiteaParams struct {
	AppName            string
	GiteaUrl           string // e.g. "https://gitea.com"
	Owner              string
	Repository         string
	Branch             string
	Token              string
	EnableSubmodules   bool
	BasePaths          string
	OutputPathOverride string
}

// CloneGiteaRepository returns a bash script to clone a Gitea repository.
func CloneGiteaRepository(params CloneGiteaParams) string {
	command := "set -e;\n"

	outputPath := params.OutputPathOverride
	if outputPath == "" {
		outputPath = filepath.Join(params.BasePaths, params.AppName, "code")
	}

	giteaHost := params.GiteaUrl
	var scheme string
	if len(giteaHost) > 7 && giteaHost[:7] == "http://" {
		scheme = "http://"
		giteaHost = giteaHost[7:]
	} else if len(giteaHost) > 8 && giteaHost[:8] == "https://" {
		scheme = "https://"
		giteaHost = giteaHost[8:]
	} else {
		scheme = "https://"
	}

	repoClone := fmt.Sprintf(
		"%s/%s/%s.git",
		giteaHost,
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
	echo "📥 Cloning Gitea Repo %[6]s to %[1]s...";
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

type GiteaRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    struct {
		UserName string `json:"login"`
	} `json:"owner"`
}

type GiteaBranch struct {
	Name string `json:"name"`
}

// GetGiteaRepositories lists Gitea repositories accessible to the token owner.
func GetGiteaRepositories(
	ctx context.Context,
	accessToken string,
	giteaURL string,
) ([]GiteaRepo, error) {
	baseUrl := strings.TrimSuffix(giteaURL, "/")
	url := fmt.Sprintf("%s/api/v1/user/repos?per_page=100", baseUrl)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"gitea api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var repos []GiteaRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// GetGiteaBranches lists branches of a Gitea repository.
func GetGiteaBranches(
	ctx context.Context,
	accessToken string,
	giteaURL string,
	owner string,
	repo string,
) ([]GiteaBranch, error) {
	baseUrl := strings.TrimSuffix(giteaURL, "/")
	url := fmt.Sprintf(
		"%s/api/v1/repos/%s/%s/branches?per_page=100",
		baseUrl,
		owner,
		repo,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"gitea api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var branches []GiteaBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

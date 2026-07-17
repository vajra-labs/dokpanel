package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GetGithubInstallationToken fetches an installation access token for a GitHub App.
func GetGithubInstallationToken(
	ctx context.Context,
	appID int64,
	privateKeyPEM string,
	installationID string,
) (string, error) {
	// 1. Sign JWT
	tokenClaims := jwt.MapClaims{
		"iat": time.Now().Add(-60 * time.Second).Unix(), // Clock drift buffer
		"exp": time.Now().Add(10 * time.Minute).Unix(),
		"iss": fmt.Sprintf("%d", appID),
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parse github app private key: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)
	signedJWT, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign github app jwt: %w", err)
	}

	// 2. Fetch Installation Token from GitHub API
	url := fmt.Sprintf(
		"https://api.github.com/app/installations/%s/access_tokens",
		installationID,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+signedJWT)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("github api request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"github api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var res struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode github token response: %w", err)
	}

	return res.Token, nil
}

type CloneGithubParams struct {
	AppName            string
	Owner              string
	Repository         string
	Branch             string
	Token              string // Already resolved using GetGithubInstallationToken
	EnableSubmodules   bool
	BasePaths          string
	OutputPathOverride string
}

// CloneGithubRepository returns a bash script to clone a GitHub App repository.
func CloneGithubRepository(params CloneGithubParams) string {
	command := "set -e;\n"

	outputPath := params.OutputPathOverride
	if outputPath == "" {
		outputPath = filepath.Join(params.BasePaths, params.AppName, "code")
	}

	repoClone := fmt.Sprintf(
		"github.com/%s/%s.git",
		params.Owner,
		params.Repository,
	)
	cloneURL := fmt.Sprintf("https://oauth2:%s@%s", params.Token, repoClone)

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
	echo "📥 Cloning Repo %[6]s to %[1]s...";
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

type GithubRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

type GithubBranch struct {
	Name string `json:"name"`
}

// GetGithubRepositories lists all accessible repositories for the GitHub App installation.
func GetGithubRepositories(
	ctx context.Context,
	token string,
) ([]GithubRepo, error) {
	url := "https://api.github.com/installation/repositories?per_page=100"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"github api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var res struct {
		Repositories []GithubRepo `json:"repositories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Repositories, nil
}

// GetGithubBranches lists branches for a repository.
func GetGithubBranches(
	ctx context.Context,
	token string,
	owner string,
	repo string,
) ([]GithubBranch, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/branches?per_page=100",
		owner,
		repo,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"github api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var branches []GithubBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

// CheckUserRepositoryPermissions checks if the collaborator has write or admin permissions.
func CheckUserRepositoryPermissions(
	ctx context.Context,
	token string,
	owner string,
	repo string,
	username string,
) (bool, string, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/collaborators/%s/permission",
		owner,
		repo,
		username,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, "", nil
	}

	var res struct {
		Permission string `json:"permission"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, "", err
	}

	allowed := map[string]bool{
		"write":    true,
		"admin":    true,
		"maintain": true,
	}

	return allowed[res.Permission], res.Permission, nil
}

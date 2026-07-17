package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"
)

type CloneBitbucketParams struct {
	AppName            string
	Owner              string
	Repository         string
	Branch             string
	Username           string
	AppPassword        string
	EnableSubmodules   bool
	BasePaths          string
	OutputPathOverride string
}

// CloneBitbucketRepository returns a bash script to clone a Bitbucket repository.
func CloneBitbucketRepository(params CloneBitbucketParams) string {
	command := "set -e;\n"

	outputPath := params.OutputPathOverride
	if outputPath == "" {
		outputPath = filepath.Join(params.BasePaths, params.AppName, "code")
	}

	repoClone := fmt.Sprintf(
		"bitbucket.org/%s/%s.git",
		params.Owner,
		params.Repository,
	)
	var cloneURL string
	if params.Username != "" && params.AppPassword != "" {
		cloneURL = fmt.Sprintf(
			"https://%s:%s@%s",
			params.Username,
			params.AppPassword,
			repoClone,
		)
	} else {
		cloneURL = fmt.Sprintf("https://%s", repoClone)
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
	echo "📥 Cloning Bitbucket Repo %[6]s to %[1]s...";
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

type BitbucketRepo struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Workspace struct {
		Slug string `json:"slug"`
	} `json:"workspace"`
}

type BitbucketBranch struct {
	Name string `json:"name"`
}

// GetBitbucketRepositories lists repositories inside a Bitbucket workspace.
func GetBitbucketRepositories(
	ctx context.Context,
	username string,
	appPassword string,
	workspace string,
) ([]BitbucketRepo, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s?pagelen=100",
		workspace,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString(
		[]byte(username + ":" + appPassword),
	)
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"bitbucket api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var res struct {
		Values []BitbucketRepo `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Values, nil
}

// GetBitbucketBranches lists branches of a Bitbucket repository.
func GetBitbucketBranches(
	ctx context.Context,
	username string,
	appPassword string,
	workspace string,
	repoSlug string,
) ([]BitbucketBranch, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/%s/refs/branches?pagelen=100",
		workspace,
		repoSlug,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString(
		[]byte(username + ":" + appPassword),
	)
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"bitbucket api returned status %s: %s",
			resp.Status,
			string(body),
		)
	}

	var res struct {
		Values []BitbucketBranch `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Values, nil
}

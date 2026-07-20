package provider

import "goploy/src/db/repos"

type GitProviderDetails struct {
	Provider  repos.GitProvider
	Github    *repos.GithubProvider
	Gitlab    *repos.GitlabProvider
	Gitea     *repos.GiteaProvider
	Bitbucket *repos.BitbucketProvider
}

type GitRepository struct {
	Name string
	URL  string
}

type GitBranch struct {
	Name string
}

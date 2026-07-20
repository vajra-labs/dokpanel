package dtos

type CreateGitProviderDto struct {
	Name         string `json:"name"         validate:"required,min=1,max=255"                       doc:"Name of the Git provider"`
	ProviderType string `json:"providerType" validate:"required,oneof=GITHUB GITLAB GITEA BITBUCKET" doc:"Provider type (GITHUB | GITLAB | GITEA | BITBUCKET)" enums:"GITHUB,GITLAB,GITEA,BITBUCKET"`
	Shared       int64  `json:"shared"       validate:"required,oneof=0 1"                           doc:"1 to share with all users, 0 otherwise"`
}

type UpdateGitProviderDto struct {
	Name   string `json:"name"   validate:"required,min=1,max=255" doc:"Name of the Git provider"`
	Shared int64  `json:"shared" validate:"required,oneof=0 1"     doc:"1 to share with all users, 0 otherwise"`
}

type GitProviderResDto struct {
	ID           int64                 `json:"id"                  doc:"Git Provider ID"`
	Name         string                `json:"name"                doc:"Name of the Git provider"`
	ProviderType string                `json:"providerType"        doc:"Provider type"`
	Shared       int64                 `json:"shared"              doc:"1 if shared, 0 otherwise"`
	CreatedAt    int64                 `json:"createdAt"           doc:"Unix timestamp created"`
	UpdatedAt    int64                 `json:"updatedAt"           doc:"Unix timestamp updated"`
	Github       *GithubProviderDto    `json:"github,omitempty"    doc:"GitHub App config if configured"`
	Gitlab       *GitlabProviderDto    `json:"gitlab,omitempty"    doc:"GitLab OAuth config if configured"`
	Gitea        *GiteaProviderDto     `json:"gitea,omitempty"     doc:"Gitea OAuth config if configured"`
	Bitbucket    *BitbucketProviderDto `json:"bitbucket,omitempty" doc:"Bitbucket Config if configured"`
}

type GithubProviderDto struct {
	ID                   int64   `json:"id"`
	GithubAppName        *string `json:"githubAppName"`
	GithubAppID          *int64  `json:"githubAppId"`
	GithubClientID       *string `json:"githubClientId"`
	GithubInstallationID *string `json:"githubInstallationId"`
	IsConfigured         bool    `json:"isConfigured"`
}

type GitlabProviderDto struct {
	ID            int64   `json:"id"`
	GitlabUrl     string  `json:"gitlabUrl"`
	ApplicationID *string `json:"applicationId"`
	GroupName     *string `json:"groupName"`
	IsConfigured  bool    `json:"isConfigured"`
}

type GiteaProviderDto struct {
	ID           int64   `json:"id"`
	GiteaUrl     string  `json:"giteaUrl"`
	ClientID     *string `json:"clientId"`
	IsConfigured bool    `json:"isConfigured"`
}

type BitbucketProviderDto struct {
	ID                int64   `json:"id"`
	BitbucketUsername *string `json:"bitbucketUsername"`
	IsConfigured      bool    `json:"isConfigured"`
}

type SaveGithubDto struct {
	GitProviderID        int64   `json:"gitProviderId"        validate:"required"`
	GithubAppName        *string `json:"githubAppName"        validate:"omitempty,max=255"`
	GithubAppID          *int64  `json:"githubAppId"          validate:"omitempty"`
	GithubClientID       *string `json:"githubClientId"       validate:"omitempty,max=255"`
	GithubClientSecret   *string `json:"githubClientSecret"   validate:"omitempty,max=1000"`
	GithubInstallationID *string `json:"githubInstallationId" validate:"omitempty,max=255"`
	GithubPrivateKey     *string `json:"githubPrivateKey"     validate:"omitempty"`
	GithubWebhookSecret  *string `json:"githubWebhookSecret"  validate:"omitempty,max=1000"`
}

type SaveGitlabDto struct {
	GitProviderID     int64   `json:"gitProviderId"     validate:"required"`
	GitlabUrl         string  `json:"gitlabUrl"         validate:"required,url"`
	GitlabInternalUrl *string `json:"gitlabInternalUrl" validate:"omitempty,url"`
	ApplicationID     *string `json:"applicationId"     validate:"omitempty,max=255"`
	RedirectUri       *string `json:"redirectUri"       validate:"omitempty,url"`
	Secret            *string `json:"secret"            validate:"omitempty,max=1000"`
	GroupName         *string `json:"groupName"         validate:"omitempty,max=255"`
}

type SaveGiteaDto struct {
	GitProviderID    int64   `json:"gitProviderId"    validate:"required"`
	GiteaUrl         string  `json:"giteaUrl"         validate:"required,url"`
	GiteaInternalUrl *string `json:"giteaInternalUrl" validate:"omitempty,url"`
	RedirectUri      *string `json:"redirectUri"      validate:"omitempty,url"`
	ClientID         *string `json:"clientId"         validate:"omitempty,max=255"`
	ClientSecret     *string `json:"clientSecret"     validate:"omitempty,max=1000"`
}

type SaveBitbucketDto struct {
	GitProviderID          int64   `json:"gitProviderId"          validate:"required"`
	BitbucketUsername      *string `json:"bitbucketUsername"      validate:"omitempty,max=255"`
	BitbucketEmail         *string `json:"bitbucketEmail"         validate:"omitempty,email"`
	AppPassword            *string `json:"appPassword"            validate:"omitempty,max=1000"`
	ApiToken               *string `json:"apiToken"               validate:"omitempty,max=1000"`
	BitbucketWorkspaceName *string `json:"bitbucketWorkspaceName" validate:"omitempty,max=255"`
}

type GitRepoDto struct {
	Name string `json:"name" doc:"Name of the repository"`
	URL  string `json:"url"  doc:"Clone URL of the repository"`
}

type GitBranchDto struct {
	Name string `json:"name" doc:"Name of the branch"`
}

package provisioner

import (
	"errors"
	"fmt"
	"net/http"

	"code.gitea.io/sdk/gitea"

	"raffy23/keptn-gitea-provisioner/pkg/keptn"
	"raffy23/keptn-gitea-provisioner/pkg/utils"
)

const DefaultPasswordLength = 32

type GiteaProvisioner struct {
	endpoint        string
	credentials     gitea.ClientOption
	client          *gitea.Client
	UsernamePrefix  string
	UserEmailDomain string
	ProjectPrefix   string
	TokenPrefix     string
}

func NewGiteaProvisioner(giteaEndpoint string, adminUsername string, adminPassword string) (*GiteaProvisioner, error) {
	clientCredentials := gitea.SetBasicAuth(adminUsername, adminPassword)
	giteaClient, err := gitea.NewClient(giteaEndpoint, clientCredentials)
	if err != nil {
		return nil, fmt.Errorf("unable to create Gitea Client: %w", err)
	}

	return &GiteaProvisioner{
		endpoint:        giteaEndpoint,
		credentials:     clientCredentials,
		client:          giteaClient,
		UsernamePrefix:  "user-",
		UserEmailDomain: "auto-provisioner.domain",
		ProjectPrefix:   "",
		TokenPrefix:     "repository-",
	}, nil
}

func (h *GiteaProvisioner) CreateUser(request *keptn.ProvisionRequest) (string, error) {

	// Generate a user:
	username := h.GetUsername(request)
	password := utils.GenerateRandomString(DefaultPasswordLength)

	// Check if user
	user, r, err := h.client.GetUserInfo(username)
	if err != nil && r == nil {
		return "", fmt.Errorf("unable to get user info for user %s: %w", username, err)
	}

	// If no user was found, we have to create the user
	if user == nil || r.StatusCode == http.StatusNotFound {
		passwordChangePolicy := false

		_, r, err := h.client.AdminCreateUser(gitea.CreateUserOption{
			LoginName:          username,
			Username:           username,
			FullName:           username,
			Email:              fmt.Sprintf("%s@%s", username, h.UserEmailDomain),
			Password:           password,
			MustChangePassword: &passwordChangePolicy,
			SendNotify:         false,
		})

		if err != nil && r == nil {
			return "", fmt.Errorf("unable to create user %s: %w", username, err)
		}
	}

	return username, nil
}

func (h *GiteaProvisioner) CreateToken(request *keptn.ProvisionRequest) (string, error) {
	// Note: we must change the client to use a different user:
	userClient, err := gitea.NewClient(h.endpoint, h.credentials, gitea.SetSudo(h.GetUsername(request)))
	if err != nil {
		return "", fmt.Errorf("unable to create gitea client: %w", err)
	}

	token, r, err := userClient.CreateAccessToken(gitea.CreateAccessTokenOption{
		Name: h.GetAccessTokenName(request),
	})
	if err != nil {
		return "", fmt.Errorf("unable to create access token: %w", err)
	}

	if r.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("recieved unkown http status code: %d", r.StatusCode)
	}

	return token.Token, nil
}

func (h *GiteaProvisioner) CreateRepository(request *keptn.ProvisionRequest) (string, error) {
	projectName := h.GetProjectName(request)
	projectDesc := fmt.Sprintf(
		"Repository was automatically provisioned by keptn-gitea-provisioner for project %s",
		projectName,
	)

	repo, r, err := h.client.AdminCreateRepo(h.GetUsername(request), gitea.CreateRepoOption{
		Name:          projectName,
		Description:   projectDesc,
		Private:       true,
		IssueLabels:   "",
		AutoInit:      false,
		Template:      false,
		Gitignores:    "",
		License:       "",
		Readme:        "",
		DefaultBranch: "master",
		TrustModel:    gitea.TrustModelDefault,
	})

	// Error while talking to gitea, upstream failed or something else
	if err != nil && r == nil {
		return "", fmt.Errorf("unable to create project %s: %w", projectName, err)
	}

	// Project already exists, relay the status code only
	if r.StatusCode == http.StatusConflict {
		return "", ErrRepositoryAlreadyExists
	}

	if r.StatusCode != 201 {
		return "", fmt.Errorf(
			"recieved unexpected status code %d while creating repository %s", r.StatusCode, request.Project,
		)
	}

	return repo.CloneURL, nil
}

func (h *GiteaProvisioner) GetUsername(request *keptn.ProvisionRequest) string {
	return fmt.Sprintf("%s%s", h.UsernamePrefix, request.Namespace)
}

func (h *GiteaProvisioner) GetProjectName(request *keptn.ProvisionRequest) string {
	return fmt.Sprintf("%s%s", h.ProjectPrefix, request.Project)
}

func (h *GiteaProvisioner) GetAccessTokenName(request *keptn.ProvisionRequest) string {
	return fmt.Sprintf("%s%s", h.TokenPrefix, request.Project)
}

func (h *GiteaProvisioner) DeleteRepository(request *keptn.ProvisionRequest) error {

	username := h.GetUsername(request)
	accessToken := h.GetAccessTokenName(request)

	r, err := h.client.DeleteRepo(username, request.Project)
	if err != nil && r == nil {
		return fmt.Errorf("unable to delete the repository: %w", err)
	}

	// Project does not exist, relay the status code only
	if r.StatusCode == http.StatusNotFound {
		return ErrRepositoryDoesNotExist
	}

	// Note: to delete a access token we have to use sudo mode:
	userClient, err := gitea.NewClient(h.endpoint, h.credentials, gitea.SetSudo(username))
	if err != nil {
		return fmt.Errorf("unable create gitea client: %w", err)
	}

	_, err = userClient.DeleteAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("unable to delete the access token: ")
	}

	return nil
}

func (h *GiteaProvisioner) ProvisionRepository(request *keptn.ProvisionRequest) (*keptn.ProvisionResponse, error) {

	if _, err := h.CreateUser(request); err != nil {
		return nil, fmt.Errorf("Unable to create user: %s\n", err.Error())
	}

	repository, err := h.CreateRepository(request)
	if err != nil {

		if errors.Is(err, ErrRepositoryAlreadyExists) {
			return nil, ErrRepositoryAlreadyExists
		}

		return nil, fmt.Errorf("unable to create repository: %w", err)
	}

	username := h.GetUsername(request)
	token, err := h.CreateToken(request)
	if err != nil {
		return nil, fmt.Errorf("unable to create token: %w", err)
	}

	return &keptn.ProvisionResponse{
		GitRemoteURL: repository,
		GitToken:     token,
		GitUser:      username,
	}, nil
}

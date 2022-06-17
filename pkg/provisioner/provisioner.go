package provisioner

import (
	"errors"
	"raffy23/keptn-gitea-provisioner/pkg/keptn"
)

var /*const*/ ErrRepositoryAlreadyExists = errors.New("the repository already exists")
var /*const*/ ErrRepositoryDoesNotExist = errors.New("the repository does not exist")

type Provisioner interface {

	// GetUsername returns the username that will be used for the given request
	GetUsername(request *keptn.ProvisionRequest) string

	// GetProjectName return the project name that will be used for the given request
	GetProjectName(request *keptn.ProvisionRequest) string

	// GetAccessTokenName returns the access token name that will be used for the given request
	GetAccessTokenName(request *keptn.ProvisionRequest) string

	// CreateUser creates a user for the given request and returns the username
	CreateUser(request *keptn.ProvisionRequest) (string, error)

	// CreateToken creates a secret token for the request
	CreateToken(request *keptn.ProvisionRequest) (string, error)

	// CreateRepository creates the repository for the given request
	CreateRepository(request *keptn.ProvisionRequest) (string, error)

	// DeleteRepository deletes the repository and all associated resources (e.g.: token)
	DeleteRepository(request *keptn.ProvisionRequest) error

	// ProvisionRepository creates all required resources for the given request
	ProvisionRepository(request *keptn.ProvisionRequest) (*keptn.ProvisionResponse, error)
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"

	"raffy23/keptn-gitea-provisioner/pkg/keptn"
	"raffy23/keptn-gitea-provisioner/pkg/provisioner"
)

var /*const*/ env envConfig

type envConfig struct {
	// Port on which the provisioner listens on
	Port          int    `envconfig:"RCV_PORT" default:"8080"`
	GiteaEndpoint string `envconfig:"GITEA_ENDPOINT" required:"true"`
	GiteaUser     string `envconfig:"GITEA_USER" required:"true"`
	GiteaPassword string `envconfig:"GITEA_PASSWORD" required:"true"`
}

func main() {
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	repoProvisioner, err := provisioner.NewGiteaProvisioner(env.GiteaEndpoint, env.GiteaUser, env.GiteaPassword)
	if err != nil {
		log.Fatalf("Unable to create gitea provisioner: %s", err)
	}

	http.HandleFunc("/repository", func(writer http.ResponseWriter, request *http.Request) {
		handleProvisionRepoRequest(repoProvisioner, writer, request)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", env.Port), nil); err != nil {
		log.Fatalf("Failed to serve endpoint: %s", err)
	}
}

func handleProvisionRepoRequest(repoProvisioner provisioner.Provisioner, w http.ResponseWriter, req *http.Request) {

	decodeRequestBody := func() (*keptn.ProvisionRequest, error) {
		request := new(keptn.ProvisionRequest)

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&request)
		if err != nil {
			return nil, fmt.Errorf("encountered error while decoding request body: %w")
		}

		return request, nil
	}

	switch req.Method {
	case http.MethodPost:

		request, err := decodeRequestBody()
		if err != nil {
			log.Printf("Unable to process request body: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response, err := repoProvisioner.ProvisionRepository(request)
		if err != nil {
			if errors.Is(err, provisioner.ErrRepositoryAlreadyExists) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			log.Printf("Unable to create repository: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		responseJson, err := json.Marshal(response)
		if err != nil {
			log.Printf("Unable to marshal reponse: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(responseJson)
		break

	case http.MethodDelete:

		request, err := decodeRequestBody()
		if err != nil {
			log.Printf("Unable to process request body: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = repoProvisioner.DeleteRepository(request)
		if err != nil {
			if errors.Is(err, provisioner.ErrRepositoryDoesNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			log.Printf("Unable to delete repository: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		break

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)

	}
}

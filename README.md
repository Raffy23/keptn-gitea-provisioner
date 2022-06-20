# keptn-gitea-provisioner

This repository contains a PoC for a Keptn service that is able to auto-provision git repositories in Gitea when a project 
in the Keptn bridge / Keptn CLI is created. 

## Installation

- Install Gitea in the Kubernetes cluster
  - Configure the admin credentials (e.g.: [admin-credentials.yaml](gitea/admin-credentials.yaml))in a Kubernetes secret 
  - Installation can be done by using the [install.sh](gitea/install.sh) script provided in the gitea folder
- Deploy keptn-gitea-provisioner via `skaffold dev`

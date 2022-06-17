package keptn

type ProvisionRequest struct {
	Project   string `json:"project"`
	Namespace string `json:"namespace"`
}

type ProvisionResponse struct {
	GitRemoteURL string `json:"gitRemoteURL"`
	GitToken     string `json:"gitToken"`
	GitUser      string `json:"gitUser"`
}

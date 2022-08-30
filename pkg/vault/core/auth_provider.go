package core

type AuthProvider interface {
	Authenticate(core API) (*AuthResponse, error)
}

type AuthResponse struct {
	Auth AuthData `json:"auth"`
}

type AuthData struct {
	ClientToken   string       `json:"client_token"`
	Accessor      string       `json:"accessor"`
	Policies      []string     `json:"policies"`
	LeaseDuration int          `json:"lease_duration"`
	Renewable     bool         `json:"renewable"`
	Metadata      AuthMetadata `json:"metadata"`
}

type AuthMetadata struct {
	Role                     string `json:"role"`
	ServiceAccountName       string `json:"service_account_name"`
	ServiceAccountNamespace  string `json:"service_account_namespace"`
	ServiceAccountSecretName string `json:"service_account_secret_name"`
	ServiceAccountUID        string `json:"service_account_uid"`
}

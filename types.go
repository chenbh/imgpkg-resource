package resource

type Source struct {
	Repository  string `json:"repository"`
	Tag         string `json:"tag"`
	PreReleases bool   `json:"prereleases"`

	Username string `json:"username"`
	Password string `json:"password"`

	Insecure bool     `json:"insecure"`
	CaCerts  []string `json:"ca_certs,omitempty"`
}

type Version struct {
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

type CheckRequest struct {
	Source  Source   `json:"source"`
	Version *Version `json:"version"`
}

type CheckResponse []Version

type InRequest struct {
	Source  Source   `json:"source"`
	Params  InParams `json:"params"`
	Version Version  `json:"version"`
}

type InParams struct {
}

type InResponse struct {
	Version Version `json:"version"`
}

type OutRequest struct {
	Source Source    `json:"source"`
	Params OutParams `json:"params"`
}

type OutParams struct {
	Tag  string `json:"tag"`
	Path string `json:"path"`
}

type OutResponse struct {
	Version Version `json:"version"`
}

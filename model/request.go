package model

type Request struct {
	Path    []string          `json:"path"`
	Field   map[string]string `json:"field, omitempty"`
	Force   *bool             `json:"force, omitempty"`
	IsGroup bool              `json:"is_group"`
}

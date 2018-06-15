package model

const (
	RequestParamUUID     = "UUID"
	RequestItemTypeGroup = 'G'
	RequestItemTypeEntry = 'E'
)

const (
	FIELD_TITLE    = "title"
	FIELD_USERNAME = "username"
	FIELD_PASSWORD = "password"
	FIELD_URL      = "url"
	FIELD_NOTES    = "notes"
)

type Request struct {
	Path    []string          `json:"path"`
	UUID    string            `json:"uuid"`
	Field   map[string]string `json:"field, omitempty"`
	Force   *bool             `json:"force, omitempty"`
	IsGroup bool              `json:"is_group"`
}

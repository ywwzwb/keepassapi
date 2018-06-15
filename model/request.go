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
	FIELD_NOTES    = "note"
)

type Request struct {
	Field   map[string]string `json:"field, omitempty"`
	IsGroup bool              `json:"is_group"`
}

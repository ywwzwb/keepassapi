package model

const (
	ItemTypeGroup = 0
	ItemTypeEntry = 1
)

type GroupInfo struct {
	Title string `json:"title"`
	Type  int    `json:"type"`
}

func NewGroupInfo(title string) GroupInfo {
	return GroupInfo{title, ItemTypeGroup}
}

type EntryBasicInfo struct {
	Title string `json:"title"`
	Type  int    `json:"type"`
}

func NewEntryBasicInfo(title string) EntryBasicInfo {
	return EntryBasicInfo{title, ItemTypeEntry}
}

type EntryInfo struct {
	Title    string `json:"title"`
	Type     int    `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	Note     string `json:"note"`
}

func NewEntryInfo(title string, username string, password string, note string) EntryInfo {
	return EntryInfo{title, ItemTypeEntry, username, password, note}
}

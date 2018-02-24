package model

type GroupInfo struct {
	Title string `json:"title"`
}

func NewGroupInfo(title string) GroupInfo {
	return GroupInfo{title}
}

type EntryInfo struct {
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	Note     string `json:"note"`
}

func NewEntryInfo(title string, username string, password string, note string) EntryInfo {
	return EntryInfo{title, username, password, note}
}

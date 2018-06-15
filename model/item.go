package model

import (
	"encoding/base64"

	"github.com/tobischo/gokeepasslib"
)

const (
	ItemTypeGroup = 0
	ItemTypeEntry = 1
)

type GroupInfo struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Type  int    `json:"type"`
}

func NewGroupInfo(uuid, title string) GroupInfo {
	return GroupInfo{uuid, title, ItemTypeGroup}
}
func NewGroupFromKeepassGroup(group gokeepasslib.Group) GroupInfo {
	return NewGroupInfo(itemTypeStrFromType(ItemTypeGroup)+base64.StdEncoding.EncodeToString(group.UUID[:]), group.Name)
}

// EntryBasicInfo basic entry info
type EntryBasicInfo struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Type  int    `json:"type"`
}

// NewEntryBasicInfo create item
func NewEntryBasicInfo(UDID string, title string) EntryBasicInfo {
	return EntryBasicInfo{UDID, title, ItemTypeEntry}
}

// NewEntryBasicFromKeepassEntry create item from gokeepass.entry
func NewEntryBasicFromKeepassEntry(entry gokeepasslib.Entry) EntryBasicInfo {
	return NewEntryBasicInfo(itemTypeStrFromType(ItemTypeEntry)+base64.StdEncoding.EncodeToString(entry.UUID[:]), entry.GetTitle())
}

type EntryDetailInfo struct {
	UUID     string `json:"uuid"`
	Title    string `json:"title"`
	Type     int    `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	Note     string `json:"note"`
	URL      string `json:"url"`
}

func NewEntryDetailInfo(UDID string, title string, username string, password string, note string, url string) EntryDetailInfo {
	return EntryDetailInfo{UDID, title, ItemTypeEntry, username, password, note, url}
}
func NewEntryDetailFromKeepassEntry(entry gokeepasslib.Entry) EntryDetailInfo {

	return NewEntryDetailInfo(itemTypeStrFromType(ItemTypeEntry)+base64.StdEncoding.EncodeToString(entry.UUID[:]), entry.GetTitle(), entry.GetContent("UserName"), entry.GetPassword(), entry.GetContent("Notes"), entry.GetContent("URL"))
}

func itemTypeStrFromType(t int) string {
	switch t {
	case ItemTypeEntry:
		return string(RequestItemTypeEntry)
	case ItemTypeGroup:
		return string(RequestItemTypeGroup)
	default:
		return ""
	}
}

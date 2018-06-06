package handler

import (
	"encoding/json"
	"keepassapi/helper"
	"keepassapi/model"
	"net/http"
)

func Read(w http.ResponseWriter, r *http.Request) {
	var requestInfo model.Request
	if err := json.NewDecoder(r.Body).Decode(&requestInfo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		model.NewGeneralError(http.StatusBadRequest, "json decode error: "+err.Error()).WriteIn(w)
		return
	}
	force := false
	if requestInfo.Force != nil {
		force = *requestInfo.Force
	}
	if force {
		if err := helper.SharedKeepassHelper().ReUnlock(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			err.WriteIn(w)
			return
		}
	}
	_ = force
	// path := "/" + mux.Vars(r)["path"]
	// forceVal := r.URL.Query()["force"]
	// force := false
	// if len(forceVal) > 0 && forceVal[0] == "true" {
	// 	force = true
	// }
	// if force {
	// 	helper.SharedKeepassHelper().ReUnlock()
	// }
	// group, entry, err := helper.SharedKeepassHelper().GetGroupOrEntryAtPath(path)
	// if group == nil && entry == nil {
	// 	if err != nil {
	// 		err.WriteIn(w)
	// 		return
	// 	}
	// 	model.NewGeneralError(http.StatusInternalServerError, "未知错误")
	// 	return
	// }
	// result := map[string]interface{}{}
	// if group != nil {
	// 	result["item"] = model.NewGroupInfo(group.Name)
	// 	subGroups := make([]model.GroupInfo, len(group.Groups))
	// 	for i, subGroup := range group.Groups {
	// 		subGroups[i] = model.NewGroupInfo(subGroup.Name)
	// 	}
	// 	subEntries := make([]model.EntryBasicInfo, len(group.Entries))
	// 	for i, subEntry := range group.Entries {
	// 		subEntries[i] = model.NewEntryBasicInfo(subEntry.GetTitle())
	// 	}
	// 	result["subGroups"] = subGroups
	// 	result["subEntries"] = subEntries
	// } else {
	// 	result["item"] = model.NewEntryInfo(entry.GetTitle(), entry.GetContent("UserName"), entry.GetPassword(), entry.GetContent("Notes"))
	// }
	// successResult := model.NewSuccessResult(result)
	// json.NewEncoder(w).Encode(successResult)
}

package handler

import (
	"encoding/json"
	"keepassapi/helper"
	"keepassapi/model"
	"net/http"

	"github.com/gorilla/mux"
)

func Get(w http.ResponseWriter, r *http.Request) {
	path := "/" + mux.Vars(r)["path"]
	// 先尝试列举
	group, entry, err := helper.SharedKeepassHelper().GetGroupOrEntryAtPath(path)
	if group == nil && entry == nil {
		if err != nil {
			err.WriteIn(w)
			return
		}
		model.NewGeneralError(http.StatusInternalServerError, "未知错误")
		return
	}
	result := map[string]interface{}{}
	if group != nil {
		result["item"] = model.NewGroupInfo(group.Name)
		subGroups := make([]model.GroupInfo, len(group.Groups))
		for i, subGroup := range group.Groups {
			subGroups[i] = model.NewGroupInfo(subGroup.Name)
		}
		subEntries := make([]model.EntryBasicInfo, len(group.Entries))
		for i, subEntry := range group.Entries {
			subEntries[i] = model.NewEntryBasicInfo(subEntry.GetTitle())
		}
		result["subGroups"] = subGroups
		result["subEntries"] = subEntries
	} else {
		result["item"] = []model.EntryInfo{model.NewEntryInfo(entry.GetTitle(), entry.GetContent("UserName"), entry.GetPassword(), entry.GetContent("Notes"))}
	}
	successResult := model.NewSuccessResult(result)
	json.NewEncoder(w).Encode(successResult)
	// groupResult, err := helper.SharedKeepassHelper().List(path)
	// if err == nil {
	// 	json.NewEncoder(w).Encode(groupResult)
	// 	return
	// }
	// if err.Code != helper.KEEPASS_ERROR_PATH_UNREACHABLE {
	// 	err.WriteIn(w)
	// 	return
	// }
	// // 再尝试获取指定对象

}

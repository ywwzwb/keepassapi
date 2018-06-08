package handler

import (
	"encoding/json"
	"keepassapi/helper"
	"keepassapi/model"
	"net/http"
)

// ReadDB will read group or entry from db
func ReadDB(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var requestInfo model.Request
	if helper.IsReqeustBodyEmpty(r) {
		// 兼容从url 参数列表取值
		if val, ok := r.URL.Query()["val"]; ok && len(val) > 0 {
			if err := json.Unmarshal([]byte(val[0]), &requestInfo); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				model.NewGeneralError(http.StatusBadRequest, "json decode error: "+err.Error()).WriteIn(w)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			model.NewGeneralError(http.StatusBadRequest, "缺少参数 val").WriteIn(w)
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&requestInfo); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			model.NewGeneralError(http.StatusBadRequest, "json decode error: "+err.Error()).WriteIn(w)
			return
		}
	}
	if requestInfo.Force != nil && *requestInfo.Force {
		if err := helper.SharedKeepassHelper().ReUnlock(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			err.WriteIn(w)
			return
		}
	}
	if requestInfo.IsGroup {
		parentGroup, err := helper.SharedKeepassHelper().GetGroupOfPath(requestInfo.Path)
		if err != nil || parentGroup == nil {
			w.WriteHeader(http.StatusBadRequest)
			err.WriteIn(w)
			return
		}
		result := map[string]interface{}{}
		result["group"] = model.NewGroupFromKeepassGroup(*parentGroup)
		subGroups := make([]model.GroupInfo, 0)
		for _, subGroup := range parentGroup.Groups {
			subGroups = append(subGroups, model.NewGroupFromKeepassGroup(subGroup))
		}
		result["child_groups"] = subGroups

		subEntries := make([]model.EntryBasicInfo, 0)
		for _, subEntry := range parentGroup.Entries {
			subEntries = append(subEntries, model.NewEntryBasicFromKeepassEntry(subEntry))
		}
		result["child_enties"] = subEntries
		successResult := model.NewSuccessResult(result)
		json.NewEncoder(w).Encode(successResult)
	} else {
		if entry, err := helper.SharedKeepassHelper().GetEntryFromPath(requestInfo.Path); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err.WriteIn(w)
		} else {
			result := map[string]interface{}{}
			result["entry"] = model.NewEntryDetailFromKeepassEntry(*entry)
			successResult := model.NewSuccessResult(result)
			json.NewEncoder(w).Encode(successResult)
		}
	}
}

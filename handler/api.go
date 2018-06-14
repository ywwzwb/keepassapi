package handler

import (
	"encoding/json"
	"keepassapi/helper"
	"keepassapi/model"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// ReadDB will read group or entry from db
func ReadDB(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uuid, ok := mux.Vars(r)[model.RequestParamUUID]
	if !ok || len(uuid) <= 1 {
		// uuid 为空时, 则默认显示根组
		uuid = string(model.RequestItemTypeGroup)
	}
	uuidtype := uuid[0]
	uuidbase64str := uuid[1:]
	force := false
	if forceStrArr, ok := r.URL.Query()["force"]; ok && len(forceStrArr) > 0 && (strings.ToLower(forceStrArr[0]) == "true" || forceStrArr[0] == "1") {
		force = true
	}
	if uuidtype == model.RequestItemTypeGroup {
		group, err := helper.SharedKeepassHelper().GetGroupOfUUID(uuidbase64str, force)
		if err != nil {
			if err.Code == helper.KEEPASS_ERROR_UUID_NOT_FOUND {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			err.WriteIn(w)
		}
		result := map[string]interface{}{}
		result["group"] = model.NewGroupFromKeepassGroup(*group)
		subGroups := make([]model.GroupInfo, 0)
		for _, subGroup := range group.Groups {
			subGroups = append(subGroups, model.NewGroupFromKeepassGroup(subGroup))
		}
		result["child_groups"] = subGroups

		subEntries := make([]model.EntryBasicInfo, 0)
		for _, subEntry := range group.Entries {
			subEntries = append(subEntries, model.NewEntryBasicFromKeepassEntry(subEntry))
		}
		result["child_enties"] = subEntries
		successResult := model.NewSuccessResult(result)
		json.NewEncoder(w).Encode(successResult)
	} else if uuidtype == model.RequestItemTypeGroup {

	} else {
		w.WriteHeader(http.StatusBadRequest)
		model.NewGeneralError(http.StatusBadRequest, "uuid 不正确")
		return
	}
	return
	// var requestInfo model.Request
	// if helper.IsReqeustBodyEmpty(r) {
	// 	// 兼容从url 参数列表取值
	// 	if val, ok := r.URL.Query()["val"]; ok && len(val) > 0 {
	// 		if err := json.Unmarshal([]byte(val[0]), &requestInfo); err != nil {
	// 			w.WriteHeader(http.StatusBadRequest)
	// 			model.NewGeneralError(http.StatusBadRequest, "json decode error: "+err.Error()).WriteIn(w)
	// 		}
	// 	} else {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		model.NewGeneralError(http.StatusBadRequest, "缺少参数 val").WriteIn(w)
	// 	}
	// } else {
	// 	if err := json.NewDecoder(r.Body).Decode(&requestInfo); err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		model.NewGeneralError(http.StatusBadRequest, "json decode error: "+err.Error()).WriteIn(w)
	// 		return
	// 	}
	// }
	// if requestInfo.Force != nil && *requestInfo.Force {
	// 	if err := helper.SharedKeepassHelper().ReUnlock(); err != nil {
	// 		w.WriteHeader(http.StatusUnauthorized)
	// 		err.WriteIn(w)
	// 		return
	// 	}
	// }
	// if requestInfo.IsGroup {
	// 	parentGroup, err := helper.SharedKeepassHelper().GetGroupOfPath(requestInfo.Path)
	// 	if err != nil || parentGroup == nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		err.WriteIn(w)
	// 		return
	// 	}
	// 	result := map[string]interface{}{}
	// 	result["group"] = model.NewGroupFromKeepassGroup(*parentGroup)
	// 	subGroups := make([]model.GroupInfo, 0)
	// 	for _, subGroup := range parentGroup.Groups {
	// 		subGroups = append(subGroups, model.NewGroupFromKeepassGroup(subGroup))
	// 	}
	// 	result["child_groups"] = subGroups

	// 	subEntries := make([]model.EntryBasicInfo, 0)
	// 	for _, subEntry := range parentGroup.Entries {
	// 		subEntries = append(subEntries, model.NewEntryBasicFromKeepassEntry(subEntry))
	// 	}
	// 	result["child_enties"] = subEntries
	// 	successResult := model.NewSuccessResult(result)
	// 	json.NewEncoder(w).Encode(successResult)
	// } else {
	// 	if entry, err := helper.SharedKeepassHelper().GetEntryFromPath(requestInfo.Path); err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		err.WriteIn(w)
	// 	} else {
	// 		successResult := model.NewSuccessResult(model.NewEntryDetailFromKeepassEntry(*entry))
	// 		json.NewEncoder(w).Encode(successResult)
	// 	}
	// }
}

// AddRecord will add group or entry to db
func AddRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var requestInfo model.Request
	if err := json.NewDecoder(r.Body).Decode(&requestInfo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		model.NewGeneralError(http.StatusBadRequest, "json decode error: "+err.Error()).WriteIn(w)
		return
	}
	if len(requestInfo.Field) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		model.NewGeneralError(http.StatusBadRequest, "Field 参数为空").WriteIn(w)
		return
	}
	var uuid *string
	var err *model.GeneralError
	if requestInfo.IsGroup {
		uuid, err = helper.SharedKeepassHelper().CreateGroupInPath(requestInfo.Path, requestInfo.Field)
	} else {
		uuid, err = helper.SharedKeepassHelper().CreateEntryInPath(requestInfo.Path, requestInfo.Field)
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err.WriteIn(w)
		return
	}
	successResult := model.NewSuccessResult(map[string]string{"uuid": *uuid})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(successResult)
	return
}

// UpdateRecord will update group or entry
func UpdateRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var requestInfo model.Request
	if err := json.NewDecoder(r.Body).Decode(&requestInfo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		model.NewGeneralError(http.StatusBadRequest, "json decode error: "+err.Error()).WriteIn(w)
		return
	}
	if len(requestInfo.Field) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var err *model.GeneralError
	if requestInfo.IsGroup {
		err = helper.SharedKeepassHelper().UpdateGroupInPath(requestInfo.Path, requestInfo.Field)
	} else {
		err = helper.SharedKeepassHelper().UpdateEntryInPath(requestInfo.Path, requestInfo.Field)
	}
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		err.WriteIn(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

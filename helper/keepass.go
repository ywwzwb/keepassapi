package helper

import (
	"encoding/base64"
	"keepassapi/model"
	"os"

	"github.com/tobischo/gokeepasslib"
)

var Keepassdbpath = ""

const (
	KEEPASS_ERROR_FILE_OPEN_FAIL = 1000 + iota
	KEEPASS_ERROR_WRONG_PASSWORD
	KEEPASS_ERROR_UNLOCK_ERROR
	KEEPASS_ERROR_DB_LOCKED
	KEEPASS_ERROR_PATH_UNREACHABLE
	KEEPASS_ERROR_PATH_UNSUPPORT
	KEEPASS_ERROR_ENCODE_ERROR
	KEEPASS_ERROR_NO_TITLE
	KEEPASS_ERROR_WRONG_UUID
	KEEPASS_ERROR_TRY_DELETE_ROOT_GROUP
	KEEPASS_ERROR_UUID_NOT_FOUND
	KEEPASS_ERROR_UNKNOWN
)
const KEEPASS_UUID_LEN = 16

type KeepassHelper struct {
	key      string
	db       *gokeepasslib.Database
	unlocked bool
}

var instance = &KeepassHelper{"", gokeepasslib.NewDatabase(), false}

// SharedKeepassHelper return the shared instance
func SharedKeepassHelper() *KeepassHelper {
	return instance
}

// TryUnlock will try to unload the db by using key
func (keepass *KeepassHelper) TryUnlock(key string) *model.GeneralError {
	if len(keepass.key) == 0 || !keepass.unlocked {
		// 还没有保存密码, 或者还未解锁, 需要解锁
		file, err := os.Open(Keepassdbpath)
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_FILE_OPEN_FAIL, "读取文件错误:"+err.Error())
		}
		defer file.Close()
		keepass.db.Credentials = gokeepasslib.NewPasswordCredentials(key)
		err = gokeepasslib.NewDecoder(file).Decode(keepass.db)
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_WRONG_PASSWORD, "密码错误:"+err.Error())
		}
		err = keepass.db.UnlockProtectedEntries()
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_UNLOCK_ERROR, "解锁错误:"+err.Error())
		}
		keepass.unlocked = true
		keepass.key = key
		return nil
	} else if keepass.key == key {
		// 如果解锁的密码和之前开锁的密码一致, 就不必重复解锁
		return nil
	}
	return model.NewGeneralError(KEEPASS_ERROR_WRONG_PASSWORD, "密码错误")
}

// ReUnlock will refresh the db
func (keepass *KeepassHelper) ReUnlock() *model.GeneralError {
	key := keepass.key
	keepass.unlocked = false
	return keepass.TryUnlock(key)
}

// GetGroupOfUUID will get the spcific keepass group
func (keepass *KeepassHelper) GetGroupOfUUID(UUIDBase64Str string, force bool) (*gokeepasslib.Group, *model.GeneralError) {
	if force {
		if err := keepass.ReUnlock(); err != nil {
			return nil, err
		}
	}
	if keepass.db == nil || keepass.unlocked == false {
		return nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := keepass.db.Content.Root
	if len(root.Groups) == 0 {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	rootGroup := &root.Groups[0]
	if len(UUIDBase64Str) == 0 {
		// 没有提供uuid 的时候, 返回根组
		return rootGroup, nil
	}
	uuid := gokeepasslib.NewUUID()
	if err := unmarshalKeepassUUIDTextURLSafe(&uuid, []byte(UUIDBase64Str)); err != nil {
		return nil, model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "UUID 异常:"+err.Error())
	}
	if _, group := keepass.findGroupInParentGroup(rootGroup, uuid); group != nil {
		return group, nil
	}
	return nil, model.NewGeneralError(KEEPASS_ERROR_UUID_NOT_FOUND, "找不到对象")
}

// GetEntryOfUUID will get the spcific keepass entry
func (keepass *KeepassHelper) GetEntryOfUUID(UUIDBase64Str string, force bool) (*gokeepasslib.Entry, *model.GeneralError) {
	if force {
		if err := keepass.ReUnlock(); err != nil {
			return nil, err
		}
	}
	if keepass.db == nil || keepass.unlocked == false {
		return nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := keepass.db.Content.Root
	if len(root.Groups) == 0 {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	rootGroup := &root.Groups[0]
	if len(UUIDBase64Str) == 0 {
		// 获取entry 必须提供uuid
		return nil, model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "未提供uuid")
	}
	uuid := gokeepasslib.NewUUID()
	if err := unmarshalKeepassUUIDTextURLSafe(&uuid, []byte(UUIDBase64Str)); err != nil {
		return nil, model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "UUID 异常:"+err.Error())
	}
	if _, entry := keepass.findEntryInParentGroup(rootGroup, uuid); entry != nil {
		return entry, nil
	}
	return nil, model.NewGeneralError(KEEPASS_ERROR_UUID_NOT_FOUND, "找不到对象")
}

// CreateGroupInParentGroup will create group in parentGroup
func (keepass *KeepassHelper) CreateGroupInParentGroup(parentUUIDBase64Str string, fields map[string]string) (*string, *model.GeneralError) {
	// 修改之前, 重新加载数据库, 保证数据一致
	parentGroup, err := keepass.GetGroupOfUUID(parentUUIDBase64Str, true)
	if err != nil {
		return nil, err
	}
	group := gokeepasslib.NewGroup()
	if title, ok := fields[model.FIELD_TITLE]; ok && len(title) > 0 {
		group.Name = title
	} else {
		return nil, model.NewGeneralError(KEEPASS_ERROR_NO_TITLE, "未设置标题")
	}
	parentGroup.Groups = append(parentGroup.Groups, group)
	uuid := string(model.RequestItemTypeGroup) + base64.StdEncoding.EncodeToString(group.UUID[:])
	keepass.saveDB()
	return &uuid, nil
}

// CreateEntryInParentGroup will create entry in parentGroup
func (keepass *KeepassHelper) CreateEntryInParentGroup(parentUUIDBase64Str string, fields map[string]string) (*string, *model.GeneralError) {
	// 修改之前, 重新加载数据库, 保证数据一致
	parentGroup, err := keepass.GetGroupOfUUID(parentUUIDBase64Str, true)
	if err != nil {
		return nil, err
	}
	entry := gokeepasslib.NewEntry()
	if title, ok := fields[model.FIELD_TITLE]; ok && len(title) > 0 {
		entry.Values = append(entry.Values, mkValue("Title", title))
	} else {
		return nil, model.NewGeneralError(KEEPASS_ERROR_NO_TITLE, "未设置标题")
	}

	if username, ok := fields[model.FIELD_USERNAME]; ok {
		entry.Values = append(entry.Values, mkValue("UserName", username))
	}
	if password, ok := fields[model.FIELD_PASSWORD]; ok {
		entry.Values = append(entry.Values, mkProtectedValue("Password", password))
	}
	if url, ok := fields[model.FIELD_URL]; ok {
		entry.Values = append(entry.Values, mkValue("URL", url))
	}
	if notes, ok := fields[model.FIELD_NOTES]; ok {
		entry.Values = append(entry.Values, mkValue("Notes", notes))
	}
	parentGroup.Entries = append(parentGroup.Entries, entry)
	uuid := string(model.RequestItemTypeEntry) + base64.StdEncoding.EncodeToString(entry.UUID[:])
	if err := keepass.saveDB(); err != nil {
		return nil, err
	}
	return &uuid, nil
}

// UpdateGroupOfUUID will update the specific group
func (keepass *KeepassHelper) UpdateGroupOfUUID(UUIDBase64Str string, fields map[string]string) *model.GeneralError {
	// 修改之前, 重新加载数据库, 保证数据一致
	group, err := keepass.GetGroupOfUUID(UUIDBase64Str, true)
	if err != nil {
		return err
	}
	if title, ok := fields[model.FIELD_TITLE]; ok {
		group.Name = title
	}
	return keepass.saveDB()
}

// UpdateEntryInPath will update an entry in specific path
func (keepass *KeepassHelper) UpdateEntryOfUUID(UUIDBase64Str string, fields map[string]string) *model.GeneralError {
	// 修改之前, 重新加载数据库, 保证数据一致
	entry, err := keepass.GetEntryOfUUID(UUIDBase64Str, true)
	if err != nil {
		return err
	}
	if title, ok := fields[model.FIELD_TITLE]; ok {
		value := entry.Get("Title")
		*value = mkValue("Title", title)
	}
	if username, ok := fields[model.FIELD_USERNAME]; ok {
		if value := entry.Get("UserName"); value != nil {
			*value = mkValue("UserName", username)
		} else {
			entry.Values = append(entry.Values, mkValue("UserName", username))
		}
	}
	if password, ok := fields[model.FIELD_PASSWORD]; ok {
		if value := entry.Get("Password"); value != nil {
			*value = mkProtectedValue("Password", password)
		} else {
			entry.Values = append(entry.Values, mkProtectedValue("Password", password))
		}
	}
	if url, ok := fields[model.FIELD_URL]; ok {
		if value := entry.Get("URL"); value != nil {
			*value = mkValue("URL", url)
		} else {
			entry.Values = append(entry.Values, mkValue("URL", url))
		}
	}
	if notes, ok := fields[model.FIELD_NOTES]; ok {
		if value := entry.Get("Notes"); value != nil {
			*value = mkValue("Notes", notes)
		} else {
			entry.Values = append(entry.Values, mkValue("Notes", notes))
		}
	}
	return keepass.saveDB()
}

// DeleteGroupOfUUID will delete specific group
func (keepass *KeepassHelper) DeleteGroupOfUUID(UUIDBase64Str string) *model.GeneralError {
	if err := keepass.ReUnlock(); err != nil {
		return err
	}
	if keepass.db == nil || keepass.unlocked == false {
		return model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := keepass.db.Content.Root
	if len(root.Groups) == 0 {
		return model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	if len(UUIDBase64Str) == 0 {
		// 删除数据必须要提供uuid
		return model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "未提供uuid")
	}
	uuid := gokeepasslib.NewUUID()
	if err := unmarshalKeepassUUIDTextURLSafe(&uuid, []byte(UUIDBase64Str)); err != nil {
		return model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "UUID 异常:"+err.Error())
	}
	rootGroup := &root.Groups[0]
	if rootGroup.UUID.Compare(uuid) {
		return model.NewGeneralError(KEEPASS_ERROR_TRY_DELETE_ROOT_GROUP, "不能删除根组")
	}
	parent, group := keepass.findGroupInParentGroup(rootGroup, uuid)
	if group == nil {
		return model.NewGeneralError(KEEPASS_ERROR_UUID_NOT_FOUND, "找不到对象")
	}
	newGroups := []gokeepasslib.Group{}
	for i := range parent.Groups {
		if parent.Groups[i].UUID.Compare(uuid) {
			continue
		}
		newGroups = append(newGroups, parent.Groups[i])
	}
	parent.Groups = newGroups
	return keepass.saveDB()
}

// DeleteEntryOfUUID will delete specific entry
func (keepass *KeepassHelper) DeleteEntryOfUUID(UUIDBase64Str string) *model.GeneralError {
	if err := keepass.ReUnlock(); err != nil {
		return err
	}
	if keepass.db == nil || keepass.unlocked == false {
		return model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := keepass.db.Content.Root
	if len(root.Groups) == 0 {
		return model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	if len(UUIDBase64Str) == 0 {
		// 删除数据必须要提供uuid
		return model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "未提供uuid")
	}
	uuid := gokeepasslib.NewUUID()
	if err := unmarshalKeepassUUIDTextURLSafe(&uuid, []byte(UUIDBase64Str)); err != nil {
		return model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "UUID 异常:"+err.Error())
	}
	rootGroup := &root.Groups[0]
	parent, entry := keepass.findEntryInParentGroup(rootGroup, uuid)
	if entry == nil {
		return model.NewGeneralError(KEEPASS_ERROR_UUID_NOT_FOUND, "找不到对象")
	}
	newEntries := []gokeepasslib.Entry{}
	for i := range parent.Entries {
		if parent.Entries[i].UUID.Compare(uuid) {
			continue
		}
		newEntries = append(newEntries, parent.Entries[i])
	}
	parent.Entries = newEntries
	return keepass.saveDB()
}
func (keepass *KeepassHelper) findGroupInParentGroup(parentGroup *gokeepasslib.Group, uuid gokeepasslib.UUID) (*gokeepasslib.Group, *gokeepasslib.Group) {
	if parentGroup == nil {
		return nil, nil
	}
	if len(uuid) != KEEPASS_UUID_LEN {
		return nil, nil
	}
	if parentGroup.UUID.Compare(uuid) {
		return nil, parentGroup
	}
	if len(parentGroup.Groups) == 0 {
		return nil, nil
	}
	for index := range parentGroup.Groups {
		if parent, group := keepass.findGroupInParentGroup(&parentGroup.Groups[index], uuid); group != nil {
			if parent == nil {
				return parentGroup, group
			}
			return parent, group
		}
	}
	return nil, nil
}

func (keepass *KeepassHelper) findEntryInParentGroup(parentGroup *gokeepasslib.Group, uuid gokeepasslib.UUID) (*gokeepasslib.Group, *gokeepasslib.Entry) {
	if parentGroup == nil {
		return nil, nil
	}
	if len(uuid) != KEEPASS_UUID_LEN {
		return nil, nil
	}
	for index := range parentGroup.Entries {
		entry := &parentGroup.Entries[index]
		if entry.UUID.Compare(uuid) {
			return parentGroup, entry
		}
	}
	for index := range parentGroup.Groups {
		if parentGroup, entry := keepass.findEntryInParentGroup(&parentGroup.Groups[index], uuid); entry != nil {
			return parentGroup, entry
		}
	}
	return nil, nil
}

func (keepass *KeepassHelper) saveDB() *model.GeneralError {
	// lock db
	keepass.db.LockProtectedEntries()
	keepass.unlocked = false
	file, oserr := os.Create(Keepassdbpath)
	if oserr != nil {
		keepass.ReUnlock()
		return model.NewGeneralError(KEEPASS_ERROR_FILE_OPEN_FAIL, "文件开启错误: "+oserr.Error())
	}
	defer file.Close()
	// get encoder
	encoder := gokeepasslib.NewEncoder(file)
	// encode
	if err := encoder.Encode(keepass.db); err != nil {
		keepass.ReUnlock()
		return model.NewGeneralError(KEEPASS_ERROR_ENCODE_ERROR, "编码错误: "+err.Error())
	}
	keepass.ReUnlock()
	return nil
}
func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value, Protected: true}}
}

func unmarshalKeepassUUIDTextURLSafe(u *gokeepasslib.UUID, text []byte) error {
	id := make([]byte, base64.URLEncoding.DecodedLen(len(text)))
	length, err := base64.URLEncoding.Decode(id, text)
	if err != nil {
		return err
	}
	if length != 16 {
		return gokeepasslib.ErrInvalidUUIDLength
	}
	copy((*u)[:], id[:16])
	return nil
}

package helper

import (
	"bytes"
	"encoding/base64"
	"fmt"
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
	KEEPASS_ERROR_UUID_NOT_FOUND
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
func (keepass *KeepassHelper) GetGroupOfUUID(uuidbase64str string, force bool) (*gokeepasslib.Group, *model.GeneralError) {
	if force {
		keepass.ReUnlock()
	}
	if keepass.db == nil || keepass.unlocked == false {
		return nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := keepass.db.Content.Root
	if len(root.Groups) == 0 {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	rootGroup := &root.Groups[0]
	if len(uuidbase64str) == 0 {
		// 没有提供uuid 的时候, 返回根组
		return rootGroup, nil
	}
	uuid := gokeepasslib.NewUUID()
	if err := uuid.UnmarshalText([]byte(uuidbase64str)); err != nil {
		return nil, model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "UUID 异常:"+err.Error())
	}
	if group := keepass.findGroupInParentGroup(rootGroup, uuid); group != nil {
		return group, nil
	}
	return nil, model.NewGeneralError(KEEPASS_ERROR_UUID_NOT_FOUND, "找不到对象")
}

// GetEntryOfUUID will get the spcific keepass entry
func (keepass *KeepassHelper) GetEntryOfUUID(uuidbase64str string, force bool) (*gokeepasslib.Entry, *model.GeneralError) {
	if force {
		keepass.ReUnlock()
	}
	if keepass.db == nil || keepass.unlocked == false {
		return nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := keepass.db.Content.Root
	if len(root.Groups) == 0 {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	rootGroup := &root.Groups[0]
	if len(uuidbase64str) == 0 {
		// 获取entry 必须提供uuid
		return nil, model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "未提供uuid")
	}
	uuid := gokeepasslib.NewUUID()
	if err := uuid.UnmarshalText([]byte(uuidbase64str)); err != nil {
		return nil, model.NewGeneralError(KEEPASS_ERROR_WRONG_UUID, "UUID 异常:"+err.Error())
	}
	if entry := keepass.findEntryInParentGroup(rootGroup, uuid); entry != nil {
		return entry, nil
	}
	return nil, model.NewGeneralError(KEEPASS_ERROR_UUID_NOT_FOUND, "找不到对象")
}

func (keepass *KeepassHelper) findGroupInParentGroup(parentGroup *gokeepasslib.Group, uuid gokeepasslib.UUID) *gokeepasslib.Group {
	if parentGroup == nil {
		return nil
	}
	if len(uuid) != KEEPASS_UUID_LEN {
		return nil
	}
	if parentGroup.UUID.Compare(uuid) {
		return parentGroup
	}
	if len(parentGroup.Groups) == 0 {
		return nil
	}
	for index := range parentGroup.Groups {
		if group := keepass.findGroupInParentGroup(&parentGroup.Groups[index], uuid); group != nil {
			return group
		}
	}
	return nil
}

func (keepass *KeepassHelper) findEntryInParentGroup(parentGroup *gokeepasslib.Group, uuid gokeepasslib.UUID) *gokeepasslib.Entry {
	if parentGroup == nil {
		return nil
	}
	if len(uuid) != KEEPASS_UUID_LEN {
		return nil
	}
	if len(parentGroup.Entries) == 0 {
		return nil
	}
	for index := range parentGroup.Entries {
		entry := &parentGroup.Entries[index]
		if entry.UUID.Compare(uuid) {
			return entry
		}
	}
	for index := range parentGroup.Groups {
		if entry := keepass.findEntryInParentGroup(&parentGroup.Groups[index], uuid); entry != nil {
			return entry
		}
	}
	return nil
}

// GetGroupOfPath will get the keepass group from specific path
func (keepass *KeepassHelper) GetGroupOfPath(path []string) (*gokeepasslib.Group, *model.GeneralError) {
	if keepass.db == nil || keepass.unlocked == false {
		return nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := keepass.db.Content.Root
	if len(root.Groups) == 0 {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	currentGroup := &root.Groups[0]
	if len(path) == 0 {
		return currentGroup, nil
	}
	for _, pathUUID := range path[1:] {
		uuid := make([]byte, 100)
		if length, err := base64.StdEncoding.Decode(uuid, []byte(pathUUID)); err != nil || length != KEEPASS_UUID_LEN {
			return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNSUPPORT, "路径"+pathUUID+"不正确")
		}
		uuid = uuid[:KEEPASS_UUID_LEN]
		var subGroup *gokeepasslib.Group
		// cg:= *currentGroup
		for i, _subGroup := range currentGroup.Groups {
			// fmt.Printf("group[%d]=%p, entry=[%p]\n", i, &, &subGroup)
			if bytes.Equal(uuid, _subGroup.UUID[:]) {
				subGroup = &currentGroup.Groups[i]
				break
			}
		}
		if subGroup == nil {
			return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "找不到对象")
		}
		currentGroup = subGroup
	}
	return currentGroup, nil
}

// GetEntryFromPath will return the entry from specific path
func (keepass *KeepassHelper) GetEntryFromPath(path []string) (*gokeepasslib.Entry, *model.GeneralError) {
	if len(path) < 2 {
		// 至少层级为2, 第一层为根节点, 第二层为entry 节点
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "路径错误, 路径长度必须大于2")
	}
	group, err := keepass.GetGroupOfPath(path[:len(path)-1])
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "查找组失败")
	}
	uuid := make([]byte, 100)
	enrtyUUID := path[len(path)-1]
	if length, err := base64.StdEncoding.Decode(uuid, []byte(enrtyUUID)); err != nil || length != KEEPASS_UUID_LEN {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNSUPPORT, "路径"+enrtyUUID+"不正确")
	}
	uuid = uuid[:KEEPASS_UUID_LEN]
	for i, entry := range group.Entries {
		if bytes.Equal(uuid, entry.UUID[:]) {
			return &group.Entries[i], nil
		}
	}
	return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "找不到对象")
}

// CreateGroupInPath will create a group in specific path, should return new group id
func (keepass *KeepassHelper) CreateGroupInPath(path []string, fields map[string]string) (*string, *model.GeneralError) {
	parentGroup, err := keepass.GetGroupOfPath(path)
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
	uuid := base64.StdEncoding.EncodeToString(group.UUID[:])
	keepass.saveDB()
	return &uuid, nil
}

// CreateEntryInPath will create an entry in specific path, should return new entry id
func (keepass *KeepassHelper) CreateEntryInPath(path []string, fields map[string]string) (*string, *model.GeneralError) {
	parentGroup, err := keepass.GetGroupOfPath(path)
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
	uuid := base64.StdEncoding.EncodeToString(entry.UUID[:])
	if err := keepass.saveDB(); err != nil {
		return nil, err
	}
	return &uuid, nil
}

// UpdateGroupInPath will update a group in specific path
func (keepass *KeepassHelper) UpdateGroupInPath(path []string, fields map[string]string) *model.GeneralError {
	group, err := keepass.GetGroupOfPath(path)
	g := &keepass.db.Content.Root.Groups[0].Groups[0]
	fmt.Printf("1--%p\n", group)
	fmt.Printf("2--%p\n", g)
	if err != nil {
		return err
	}
	if title, ok := fields[model.FIELD_TITLE]; ok {
		group.Name = title
	}
	return keepass.saveDB()
}

// UpdateEntryInPath will update an entry in specific path
func (keepass *KeepassHelper) UpdateEntryInPath(path []string, fields map[string]string) *model.GeneralError {
	entry, err := keepass.GetEntryFromPath(path)
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

package helper

import (
	"bytes"
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
)
const KEEPASS_UUID_LEN = 16

type KeepassHelper struct {
	key      string
	db       *gokeepasslib.Database
	unlocked bool
}

var instance = &KeepassHelper{"", gokeepasslib.NewDatabase(), false}

func SharedKeepassHelper() *KeepassHelper {
	return instance
}
func (self *KeepassHelper) TryUnlock(key string) *model.GeneralError {
	if len(self.key) == 0 || !self.unlocked {
		// 还没有保存密码, 或者还未解锁, 需要解锁
		file, err := os.Open(Keepassdbpath)
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_FILE_OPEN_FAIL, "读取文件错误:"+err.Error())
		}
		self.db.Credentials = gokeepasslib.NewPasswordCredentials(key)
		err = gokeepasslib.NewDecoder(file).Decode(self.db)
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_WRONG_PASSWORD, "密码错误:"+err.Error())
		}
		err = self.db.UnlockProtectedEntries()
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_UNLOCK_ERROR, "解锁错误:"+err.Error())
		}
		self.unlocked = true
		self.key = key
		return nil
	} else if self.key == key {
		// 如果解锁的密码和之前开锁的密码一致, 就不必重复解锁
		return nil
	}
	return model.NewGeneralError(KEEPASS_ERROR_WRONG_PASSWORD, "密码错误")
}
func (self *KeepassHelper) ReUnlock() *model.GeneralError {
	key := self.key
	self.unlocked = false
	return self.TryUnlock(key)
}

// GetGroupOfPath will get the keepass group from specific path
func (helper *KeepassHelper) GetGroupOfPath(path []string) (*gokeepasslib.Group, *model.GeneralError) {
	if helper.db == nil || helper.unlocked == false {
		return nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	root := helper.db.Content.Root
	if len(root.Groups) == 0 {
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	_ = root
	currentGroup := root.Groups[0]
	for _, pathUUID := range path[1:] {
		uuid := make([]byte, 100)
		if length, err := base64.StdEncoding.Decode(uuid, []byte(pathUUID)); err != nil || length != KEEPASS_UUID_LEN {
			return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNSUPPORT, "路径"+pathUUID+"不正确")
		}
		uuid = uuid[:KEEPASS_UUID_LEN]
		var subGroup *gokeepasslib.Group
		for _, _subGroup := range currentGroup.Groups {
			if bytes.Equal(uuid, _subGroup.UUID[:]) {
				subGroup = &_subGroup
				break
			}
		}
		if subGroup == nil {
			return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "找不到对象")
		}
		currentGroup = *subGroup
	}
	return &currentGroup, nil
}
func (helper *KeepassHelper) GetEntryFromPath(path []string) (*gokeepasslib.Entry, *model.GeneralError) {
	if len(path) < 2 {
		// 至少层级为2, 第一层为根节点, 第二层为entry 节点
		return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "路径错误, 路径长度必须大于2")
	}
	group, err := helper.GetGroupOfPath(path[:len(path)-1])
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
	for _, entry := range group.Entries {
		if bytes.Equal(uuid, entry.UUID[:]) {
			return &entry, nil
		}
	}
	return nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "找不到对象")
}

// func (self *KeepassHelper) GetGroupOrEntryAtPath(path string) (*gokeepasslib.Group, *gokeepasslib.Entry, *model.GeneralError) {
// 	if self.db == nil || self.unlocked == false {
// 		return nil, nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
// 	}
// 	if path[len(path)-1] == '/' {
// 		path = path[:len(path)-1]
// 	}
// 	if len(path) == 0 {
// 		path = "/"
// 	}
// 	if path[0] != '/' {
// 		return nil, nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNSUPPORT, "只支持绝对路径")
// 	}
// 	paths := strings.Split(path, "/")
// 	if len(paths) > 0 && paths[0] == "" {
// 		paths = paths[1:]
// 	}
// 	if len(paths) > 0 && paths[len(paths)-1] == "" {
// 		paths = paths[:len(paths)-1]
// 	}
// 	root := self.db.Content.Root
// 	if len(root.Groups) == 0 {
// 		return nil, nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
// 	}
// 	rootGroup := root.Groups[0]
// 	currentGroup := rootGroup
// 	for deep, subPath := range paths {
// 		notFind := true
// 		for _, group := range currentGroup.Groups {
// 			if group.Name == subPath {
// 				currentGroup = group
// 				notFind = false
// 				break
// 			}
// 		}
// 		if notFind {
// 			if deep == len(paths)-1 {
// 				// 查找实体
// 				for _, entry := range currentGroup.Entries {
// 					if entry.GetTitle() == subPath {
// 						return nil, &entry, nil
// 					}
// 				}
// 				// 实体查找完毕， 没有找到对应数据
// 				return nil, nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "找不到对象")
// 			}
// 		}
// 	}

// 	return &currentGroup, nil, nil
// }

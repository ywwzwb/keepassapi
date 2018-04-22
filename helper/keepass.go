package helper

import (
	"fmt"
	"keepassapi/model"
	"os"
	"strings"

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
	if len(self.key) == 0 {
		fmt.Println("reunloc3")
		file, err := os.Open(Keepassdbpath)
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_FILE_OPEN_FAIL, "读取文件错误")
		}
		self.db.Credentials = gokeepasslib.NewPasswordCredentials(key)
		err = gokeepasslib.NewDecoder(file).Decode(self.db)
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_WRONG_PASSWORD, "密码错误")
		}
		err = self.db.UnlockProtectedEntries()
		if err != nil {
			return model.NewGeneralError(KEEPASS_ERROR_UNLOCK_ERROR, "解锁错误")
		}
		self.unlocked = true
		self.key = key
		return nil
	} else if self.key == key {
		return nil
	}
	return model.NewGeneralError(KEEPASS_ERROR_WRONG_PASSWORD, "密码错误")
}
func (self *KeepassHelper) ReUnlock() *model.GeneralError {
	fmt.Println("reunlock2", self.key)
	key := self.key
	self.key = ""
	return self.TryUnlock(key)
}
func (self *KeepassHelper) GetGroupOrEntryAtPath(path string) (*gokeepasslib.Group, *gokeepasslib.Entry, *model.GeneralError) {
	if self.db == nil || self.unlocked == false {
		return nil, nil, model.NewGeneralError(KEEPASS_ERROR_DB_LOCKED, "数据库未解锁")
	}
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if len(path) == 0 {
		path = "/"
	}
	if path[0] != '/' {
		return nil, nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNSUPPORT, "只支持绝对路径")
	}
	paths := strings.Split(path, "/")
	if len(paths) > 0 && paths[0] == "" {
		paths = paths[1:]
	}
	if len(paths) > 0 && paths[len(paths)-1] == "" {
		paths = paths[:len(paths)-1]
	}
	root := self.db.Content.Root
	if len(root.Groups) == 0 {
		return nil, nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "空数据库")
	}
	rootGroup := root.Groups[0]
	currentGroup := rootGroup
	for deep, subPath := range paths {
		notFind := true
		for _, group := range currentGroup.Groups {
			if group.Name == subPath {
				currentGroup = group
				notFind = false
				break
			}
		}
		if notFind {
			if deep == len(paths)-1 {
				// 查找实体
				for _, entry := range currentGroup.Entries {
					if entry.GetTitle() == subPath {
						return nil, &entry, nil
					}
				}
				// 实体查找完毕， 没有找到对应数据
				return nil, nil, model.NewGeneralError(KEEPASS_ERROR_PATH_UNREACHABLE, "找不到对象")
			}
		}
	}

	return &currentGroup, nil, nil
}

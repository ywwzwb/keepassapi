package helper

import (
	"keepassapi/model"
	"os"

	"github.com/tobischo/gokeepasslib"
)

const keepassdbpath = "/Users/zwb/Documents/keepass.kdbx"
const (
	KEEPASS_ERROR_FILE_OPEN_FAIL = 1000 + iota
	KEEPASS_ERROR_WRONG_PASSWORD
	KEEPASS_ERROR_UNLOCK_ERROR
)

type KeepassHelper struct {
	key string
	db  *gokeepasslib.Database
}

var instance = &KeepassHelper{"", gokeepasslib.NewDatabase()}

func SharedKeepassHelper() *KeepassHelper {
	return instance
}
func (self *KeepassHelper) TryUnlock(key string) *model.GeneralError {
	if len(self.key) == 0 {
		file, err := os.Open(keepassdbpath)
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
		self.key = key
		return nil
	} else if self.key == key {
		return nil
	}
	return model.NewGeneralError(KEEPASS_ERROR_WRONG_PASSWORD, "密码错误")
}

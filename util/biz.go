package util

import (
	"encoding/base64"
	"fmt"
	"github.com/puras/mog/constants"
	"github.com/puras/mog/ctype"
	"github.com/puras/mog/errdef"
	"strings"
	"time"
)

/**
* @project momo-backend
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-10-03 16:07
 */

// GenerateSalt 生成Salt
func GenerateSalt(account string, createdAt ctype.Time) string {
	createdAtStr := time.Time(createdAt).Format("2006-01-02 15:04:05")
	return Sha256Encrypt(fmt.Sprintf("%s-%s", account, createdAtStr))
}

// GeneratePassword 生成密码
func GeneratePassword(password, salt string) string {
	return Sha256Encrypt(fmt.Sprintf("%s%s", password, salt))
}

// ParsePassword 解析密码
func ParsePassword(password string) (string, error) {
	passwdBytes, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", errdef.New(errdef.InvalidParam)
	}
	passwdStr := string(passwdBytes)
	passwdInfos := strings.Split(passwdStr, fmt.Sprintf("%s%s", constants.PasswordPrefix, constants.PasswordSep))
	if len(passwdInfos) != 2 {
		return "", errdef.New(errdef.InvalidParam)
	}
	passwdAndTimestamp := passwdInfos[1]
	if strings.Index(passwdAndTimestamp, constants.PasswordSep) == -1 {
		return "", errdef.New(errdef.InvalidParam)
	}
	return passwdAndTimestamp[:strings.LastIndex(passwdAndTimestamp, constants.PasswordSep)], nil
}

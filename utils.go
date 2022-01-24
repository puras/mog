package mog

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	uuid "github.com/satori/go.uuid"
)

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 13:07
 * @desc
 */
func GenUUID4() string {
	u4 := uuid.NewV4()
	return u4.String()
}

func GenShortUUID() string {
	ret := GenUUID4()
	return strings.ReplaceAll(ret, "-", "")
}

func IdShortString(id string, len int) string {
	sign := []byte(id)
	hash := md5.New()
	hash.Write(sign)
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))
	return hex.EncodeToString(hash.Sum(nil))[:len]
}

func GenInfoCode(prefix string, infoType string, id string) string {
	ret := fmt.Sprintf("%s%s-%s", prefix, infoType, IdShortString(id, ShortIdLen))
	return strings.ToUpper(ret)
}

func Sha256Encrypt(info string) string {
	h := sha256.New()
	h.Write([]byte(info))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func IsNil(i interface{}) bool {
	defer func() {
		recover()
	}()
	vi := reflect.ValueOf(i)
	return vi.IsNil()
}

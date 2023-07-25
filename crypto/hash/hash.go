package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func MD5(b []byte) string {
	h := md5.New()
	_, _ = h.Write(b)
	return fmt.Sprintf("%s", h.Sum(nil))
}

func MD5String(s string) string {
	return MD5([]byte(s))
}

func SHA1(b []byte) string {
	h := sha1.New()
	_, _ = h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func SHA1String(s string) string {
	return SHA1([]byte(s))
}

func GeneratePassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CompareHashAndPassword(hashedPasswd, passwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPasswd), []byte(passwd))
}

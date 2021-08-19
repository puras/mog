package util

import uuid "github.com/satori/go.uuid"

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

package utils

import "github.com/rs/xid"

/**
 * @project heqoo-go
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2023-04-12 15:05
 * @desc
 */

func NewID() string {
	return xid.New().String()
}

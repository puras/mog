package mog

import "fmt"

/**
* @project kudo
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-09-13 21:18
 */
func ReportPanic() {
	p := recover()
	if p == nil {
		return
	}
	err, ok := p.(error)
	if ok {
		fmt.Println("启动出错", err)
	}
}

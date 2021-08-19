package util

import (
	"fmt"
	"os"
)

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 21:43
 */
// PathExists 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// MakeDir 创建文件夹
func MakeDir(dir string) error {
	exist, err := PathExists(dir)
	if err != nil {
		fmt.Printf("get dir errcode![%v]\n", err)
		return err
	}
	if exist {
		fmt.Printf("has dir![%v]\n", dir)
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("mkdir failed![%v]\n", err)
	} else {
		fmt.Printf("mkdir success!\n")
	}
	return err
}

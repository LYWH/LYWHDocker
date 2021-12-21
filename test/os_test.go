package test

import (
	"LYWHDocker/container"
	"fmt"
	"testing"
)

/*
createTime:LYWH
createData:2021/12/16
*/
func TestOsStat(t *testing.T) {
	filePath := "/home/lywh/Learning/docke"
	f, err := container.DirOrFileExist(filePath)
	fmt.Println("-====", f)
	fmt.Println("++++++", err)
}

package test

import (
	"fmt"
	"os"
	"testing"
)

/*
createTime:LYWH
createData:2021/12/16
*/
func TestOsStat(t *testing.T) {
	filePath := "/home/lywh/Learning/docker/Dockerfile"
	f, err := os.Stat(filePath)
	fmt.Println("-====", f)
	fmt.Println("++++++", err)
	fmt.Println(os.IsNotExist(err))
}

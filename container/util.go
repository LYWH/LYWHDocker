package container

import (
	"LYWHDocker/log"
	//"github.com/go-kit/kit/log"
	"os"
)

/*
createTime:LYWH
createData:2021/12/20
*/
func DirOrFileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		log.Mylog.Error(err)
		return false, err
	}
}

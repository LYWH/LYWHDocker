package subsystems

/*
createTime:LYWH
createData:2021/12/16
*/

//实现子系统类型

type ResourceConfig struct {
	Memory string //对内存资源的限制
}

type SubSystem interface {
	Name() string                      //返回子系统的名字
	Set(string, *ResourceConfig) error //设置资源限制
	Apply(string, int) error           //将进程添加到cgroup中
	Remove(string) error               //移除cgroup
}

var SubSystemChains = []SubSystem{
	&MemoeySubSystem{},
}

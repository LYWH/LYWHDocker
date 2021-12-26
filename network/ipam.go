package network

/*
createTime:LYWH
createData:2021/12/26
*/

const (
	defaultIPAMAllocatorPath = "/var/run/LYWHDocker/network/ipam/subnet.json"
)

type IPAM struct {
	SubnetAllocatorPath string            //文件分配存储的位置
	Subnets             map[string]string //网段和位图算法数组的map，其中key代表网段，value代表的是分配的位图字符串（一位代表一个IP的分配标识符）
}

var ipAllocator = &IPAM{SubnetAllocatorPath: defaultIPAMAllocatorPath}

//加载ipam的位图信息
func (i *IPAM) load() error {

}

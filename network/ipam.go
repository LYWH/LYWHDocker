package network

import (
	"LYWHDocker/container"
	"LYWHDocker/log"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

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
func (ipam *IPAM) load() error {
	//检查文件是否存在
	if has, err := container.DirOrFileExist(ipam.SubnetAllocatorPath); err == nil && !has {
		//如果不存在则说明还未分配
		log.Mylog.Info("load", "container.DirOrFileExist", err)
		return nil
	}

	configFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		log.Mylog.Error("load", "os.Open", err)
		return err
	}
	//读取文件数据
	byteData, err := ioutil.ReadAll(configFile)
	err = json.Unmarshal(byteData, &ipam.Subnets) //反序列化到对象中
	if err != nil {
		log.Mylog.Error("load", "json.Unmarshal", err)
		return err
	}
	return nil
}

//将IP分配信息存储到文件中
func (ipam *IPAM) dump() error {
	//先检测存储的文件夹是否存在,不存在则创建
	configPath, _ := path.Split(ipam.SubnetAllocatorPath)
	if has, err := container.DirOrFileExist(configPath); err == nil && !has {
		log.Mylog.Error("load", "container.DirOrFileExist", err)
		if err = os.MkdirAll(configPath, 0755); err != nil {
			log.Mylog.Error("create path failed", err)
		}
		return err
	}
	//打开文件，其中os.O_TRUNC表示：若文件存在，则其长度被截断为0，相当于把原有内容给覆盖了，os.O_CREATE表示如果文件不存在则创建
	configFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer configFile.Close()
	if err != nil {
		log.Mylog.Error("open config file failed", "load", "os.OpenFile", err)
		return err
	}
	//序列化ip分配信息
	configContentJSON, err := json.Marshal(ipam.Subnets)
	if err != nil {
		log.Mylog.Error("Serialization  error", "dump", "json.Marshal", err)
		return err
	}
	if _, err = configFile.Write(configContentJSON); err != nil {
		log.Mylog.Error("write  error", "dump", "configFile.Write", err)
		return err
	}
	return nil
}

//bitmap算法，用于分配IP地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	//存放网段地址分配信息
	ipam.Subnets = make(map[string]string)
	//从文件中加载网段地址分配信息
	if err := ipam.load(); err != nil {
		log.Mylog.Error("Allocate", "ipam.load", err)
		return nil, err
	}
	//由于subnet是指针类型，为了避免影响，重新生成新的subnet
	_, subnet, _ = net.ParseCIDR(subnet.String())
	// subnet.Mask.Size() 函数会返回网段前面的固定位1的长度以及后面0位的长度
	// 例如: 127.0.0.0/8 子网掩码：255.0.0.0 subnet.Mask.Size()返回8和32
	one, size := subnet.Mask.Size()
	ipAddr := subnet.String()
	//如果之前没有分配过该网段，则初始化该网段信息
	if _, has := ipam.Subnets[ipAddr]; !has {
		//1<<unit8(num)表示2^num，下行代码表示
		ipam.Subnets[ipAddr] = strings.Repeat("0", 1<<uint8(size-one))
	}
	var allIP net.IP
	//遍历网段的位图数组
	for c := range ipam.Subnets[ipAddr] {
		if ipam.Subnets[ipAddr][c] == '0' {
			//将第一个0修改为1，对应位置的IP表示是分配的IP
			//syting类型不能直接修改，需要转换为[]byte类型
			ipBytes := []byte(ipam.Subnets[ipAddr])
			ipBytes[c] = '1'
			ipam.Subnets[ipAddr] = string(ipBytes)
			firstIP := subnet.IP //该网段的第一个IP，即网关IP
			// ip地址是一个uint[4]的数组，对于网段172.16.0.0/12，172.16.0.0就是[172, 16, 0, 0]
			// 需要通过数组中每一项加所需要的值, 对于当前序号,例如65535
			// 每一位加的计算就是[uint8(65535>>24), uint8(65535>>16), uint8(65535>>8), uint8(65535>>0)]
			//即[0,1,0,19]，因此计算后得到的IP是[172,17,0,19]
			for t := uint8(4); t > 0; t-- {
				[]byte(firstIP)[4-t] += uint8((c + 1) >> ((t - 1) * 8))
			}
			//由于网段的第一个地址是网管地址，但是由于序号+1，因此后续已经计算好的IP中不需要+1操作，避免了每个段的临界情况导致归零
			allIP = firstIP
			break
		}
	}
	//将已经分配的IP状态写回文件
	if err := ipam.dump(); err != nil {
		log.Mylog.Error("Allocate", "ipam.dump", err)
		return nil, err
	}
	return allIP, nil
}

//回收IP地址,网段和IP
func (ipam *IPAM) Release(subnet *net.IPNet, ipAddr *net.IP) error {
	//由于序列化和反序列化的需要，因此需要重新生成Subnets
	ipam.Subnets = make(map[string]string)
	//由于subnet是指针类型，为了避免影响，重新生成新的subnet
	_, subnet, _ = net.ParseCIDR(subnet.String())
	//加载ipam的网络配置信息
	if err := ipam.load(); err != nil {
		log.Mylog.Error("Release", "ipam.load", err)
		return err
	}
	//c表示ip状态表的索引位置
	c := 0
	//
	realiseIP := ipAddr.To4()
	//由于index为0表示第一个课分配的IP，在分配IP时将C加一，因此还原时需要减一
	realiseIP[3] -= 1
	for t := uint8(4); t > 0; t-- {
		c += int((realiseIP[4-t] - subnet.IP[4-t]) << (t - 1) * 8)
	}
	ipalloc := []byte(ipam.Subnets[subnet.String()])
	ipalloc[c] = '0'
	//反序列化
	ipam.Subnets[subnet.String()] = string(ipalloc)
	if err := ipam.dump(); err != nil {
		log.Mylog.Error("Release", "ipam.dump()", err)
		return err
	}
	return nil
}

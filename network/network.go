package network

import (
	"LYWHDocker/container"
	"LYWHDocker/log"
	"encoding/json"
	"github.com/vishvananda/netlink"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
)

/*
createTime:LYWH
createData:2021/12/25
*/
type NetWork struct {
	Name    string     //网络名
	IpRange *net.IPNet //网路Ip地址范围
	Driver  string     //网络驱动
}

type EndPoint struct {
	ID          string
	Device      netlink.Veth     //veth设备
	IPAddress   net.IP           //IP地址
	MacAdderss  net.HardwareAddr //mac地址
	PortMapping []string         //端口映射
	Network     *NetWork         //网络
}

type NetWorkDriver interface {
	Name() string                                          //获取驱动的名字
	Create(subnet, name string) (*NetWork, error)          //创建网络
	Delete(network *NetWork) error                         //删除网络
	Connect(netWork *NetWork, endPoint *EndPoint) error    //终端设备连接网络
	DisConnect(netWork *NetWork, endPoint *EndPoint) error //终端设备卸载网络
}

var (
	defaultNetWorkPath = "/var/run/LYWHDocker/network/network/" //
	drivers            = map[string]NetWorkDriver{}             //网络驱动映射映射
	networks           = map[string]*NetWork{}                  //网络映射
)

//初始化程序，内容：加载网络驱动到map中，加载所有网络到内存中
func Init() error {
	//加载网络驱动
	var bridgeDriver = BridgeNetWorkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	//defaultNetWorkPath不存在就创建
	if has, err := container.DirOrFileExist(defaultNetWorkPath); err == nil && !has {
		log.Mylog.Info("no file", defaultNetWorkPath)
		if err = os.MkdirAll(defaultNetWorkPath, 0755); err != nil {
			log.Mylog.Error("Init", "os.MkdirAll", err)
			return err
		}
	}
	//遍历配置文件夹下面得所有文件
	if err := filepath.Walk(defaultNetWorkPath, func(nwPath string, info fs.FileInfo, err error) error {
		//如果是文件夹则跳过
		if info.IsDir() {
			return nil
		}
		//将文件名字作为网络名
		_, networkName := path.Split(nwPath)
		netWork := &NetWork{
			Name: networkName,
		}
		//加载网络配置信息
		if err = netWork.load(nwPath); err != nil {
			log.Mylog.Error("Init", "netWork.load", err)
			return err
		}
		//加载网络配置信息到network字典中
		networks[netWork.Name] = netWork
		return nil
	}); err != nil {
		return err
	}
	return nil
}

//将配置文件加载到NetWork对象中,主要过程是读取数据和反序列化
func (nw *NetWork) load(dumpPath string) error {
	//打开文件
	configFile, err := os.Open(dumpPath)
	if err != nil {
		log.Mylog.Error("load", "os.Open", err)
		return err
	}
	defer configFile.Close()
	//读取数据
	jsonBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Mylog.Error("load", "ioutil.ReadAll", err)
		return err
	}
	//反序列化读取得数据到网络对象中
	if err = json.Unmarshal(jsonBytes, &nw); err != nil {
		log.Mylog.Error("load", "json.Unmarshal", err)
		return err
	}
	return nil
}

//根据网络驱动创建网络
func CerateNetWork(dirver, subnet, name string) error {
	//将网段字符串转换为ipnet对象
	_, ipnet, _ := net.ParseCIDR(subnet)

}

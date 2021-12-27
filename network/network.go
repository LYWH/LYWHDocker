package network

import (
	"LYWHDocker/container"
	"LYWHDocker/log"
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
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
func CreateNetWork(dirver, subnet, name string) error {
	//将网段字符串转换为ipnet对象
	_, ipnet, _ := net.ParseCIDR(subnet)
	//分配网关，并将网段中的第一个IP作为网关IP
	gatewayIP, err := ipAllocator.Allocate(ipnet)
	if err != nil {
		log.Mylog.Error("CerateNetWork", "ipAllocator.Allocate", err)
		return err
	}
	ipnet.IP = gatewayIP
	network, err := drivers[dirver].Create(ipnet.String(), name)
	if err != nil {
		log.Mylog.Error("CerateNetWork", "drivers[dirver].Create", err)
		return err
	}
	return network.dump(defaultNetWorkPath)
}

//将网络信息保存到文件中
func (network *NetWork) dump(filepath string) error {
	if has, err := container.DirOrFileExist(filepath); err == nil && !has {
		//如果没有文件就创建文件
		if err = os.MkdirAll(filepath, 0644); err != nil {
			log.Mylog.Error("dump", "os.MkdirAll", err)
			return err
		}
	}
	networkFile := path.Join(filepath, network.Name)
	file, err := os.OpenFile(networkFile, os.O_TRUNC|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Mylog.Error("dump", "os.OpenFile", err)
		return err
	}
	//序列化network
	byteData, err := json.Marshal(network)
	if err != nil {
		log.Mylog.Error("dump", "json.Marshal", err)
		return err
	}
	//将序列化内容写入文件
	if _, err = file.Write(byteData); err != nil {
		log.Mylog.Error("dump", "file.Write", err)
		return err
	}
	return nil
}

//容器连接网络
func Connect(networkName string, containerInfo *container.ContainerInfo) error {
	//networks字典中保存了网络的信息，因此可以通过名字key直接获取
	network, ok := networks[networkName]
	if !ok {
		err := fmt.Errorf("no such network %s\n", networkName)
		log.Mylog.Error(err)
		return err
	}
	//通过ipam获取网段可用IP
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		log.Mylog.Error("Connect", "ipAllocator.Allocate", err)
		return err
	}
	//创建网络端点，用于连接容器
	ep := &EndPoint{
		ID:          fmt.Sprintf("%s-%s", containerInfo.Id, networkName),
		IPAddress:   ip,
		PortMapping: containerInfo.PortMapping,
		Network:     network,
	}
	//进入容器网络中的net namespace，调用网络驱动方法连接网络和端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		log.Mylog.Error("Connect", "drivers[network.Driver].Connect", err)
		return err
	}
	//进入容器的namespace配置IP和路由
	if err := configEndPointIPAdressAndRoute(ep, containerInfo); err != nil {
		log.Mylog.Error("Connect", "configEndPointIPAdressAndRoute", err)
		return err
	}
	//配置容器的宿主机和端口映射
	return configPortMapping(ep)

}

//配置网络端点的ip和路由
func configEndPointIPAdressAndRoute(ep *EndPoint, cinfo *container.ContainerInfo) error {
	//找到veth的另外一端
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		log.Mylog.Error("configEndPointIPAdressAndRoute", "netlink.LinkByName", err)
		return err
	}
	//将端点加入到容器的net namespace中
	//同时该函数后续的内容都是在容器的net namespace中
	//函数结束后会恢复到执行进程的默认空间
	defer enterContainerNetNameSpace(&peerLink, cinfo)()

	//获取容器的IP地址以及网段
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress
	//设置容器内veth端点的IP
	if err := setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		log.Mylog.Error("configEndPointIPAdressAndRoute", "setInterfaceIP", err)
		return err
	}
	//启动容器内的端点
	if err := setInterfaceUp(ep.Device.PeerName); err != nil {
		log.Mylog.Error("configEndPointIPAdressAndRoute", "setInterfaceUp", err)
		return err
	}
	//启动容器内的lo网卡，启动后可以保证容器内访问自己的请求
	if err := setInterfaceUp("lo"); err != nil {
		log.Mylog.Error("configEndPointIPAdressAndRoute", "setInterfaceUp", err)
		return err
	}

	//设置容器内所有访问请求都经过veth的端点
	//0.0.0.0/0网段表示所有的IP段
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	//构建默认路由，包括设备、网关、网段
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		log.Mylog.Error("configEndPointIPAdressAndRoute", "netlink.RouteAdd", err)
		return err
	}
	return nil
}

//使当前线程进入net namespace，并配置veth
//锁定当前执行的线程，方式go将此线程调度执行其它goroutine而离开了此目标空间
//返回的是一个函数指针，执行这个函数时会退出容器的那net namespace，返回到宿主机的namespace
func enterContainerNetNameSpace(enLink *netlink.Link, containerInfo *container.ContainerInfo) func() {
	//通过读取文件的方式找到容器net namespace，关键参数是容器进程的ID
	file, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", containerInfo.Pid), os.O_RDONLY, 0644)
	if err != nil {
		log.Mylog.Error("enterContainerNetNameSpace", "os.OpenFile", err)
		return nil
	}
	//获取容器net namespace文件的文件描述符
	nsFD := file.Fd()
	//锁定当前进程
	runtime.LockOSThread()
	//修改veth的另外一端，将其移动到容器中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.Mylog.Error("enterContainerNetNameSpace", "netlink.LinkSetNsFd", err)
		return nil
	}
	//获取当前进程的net namespace，以便退出容器进程时回到当前的net namespace
	originNetNameSpace, err := netns.Get()
	if err != nil {
		log.Mylog.Error("enterContainerNetNameSpace", "netns.Get", err)
		return nil
	}
	//将当前进程加入到容器中，执行返回函数时才退出
	if err := netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Mylog.Error("enterContainerNetNameSpace", "netns.Set", err)
		return nil
	}
	return func() {
		//返回原先的net namespace
		netns.Set(originNetNameSpace)
		//关闭原先的文件
		originNetNameSpace.Close()
		//取消线程的锁
		runtime.UnlockOSThread()
		//关闭NameSpace文件
		file.Close()
	}
}

//配置端口映射
//容器内部又独立的网络空间和IP地址，但是该地址无法访问到宿主机的外部地址，因此需要做端口映射以便访问外部地址
func configPortMapping(ep *EndPoint) error {
	//遍历容器端口映射表
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Mylog.Error("port mapping format error\n", pm)
			continue
		}
		//试用cmd的方式向iptables的PREROUTING中添加DNAT规则,将宿主机的端口转发请求转发到容器的地址和端口上
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp -dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.Mylog.Error("iptables output\n", output)
			continue
		}
	}
	return nil
}

func ListNetWork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "Name\tIpRange\tDriver")
	//遍历网络信息
	for _, nw := range networks {
		fmt.Fprint(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange, nw.Driver)
	}
	if err := w.Flush(); err != nil {
		log.Mylog.Error("error output\n", err)
	}
}

func DeleteWork(networkName string) error {
	//检查网络是否存在
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("can't find network %s\n", networkName)
	}
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		log.Mylog.Error("DeleteWork", "ipAllocator.Release", err)
		return err
	}
	if err := drivers[nw.Driver].Delete(nw); err != nil {
		log.Mylog.Error("DeleteWork", "drivers[nw.Driver].Delete", err)
		return err
	}
	//删除对应的网络
	return nw.DeleteNetWork(defaultNetWorkPath)
}

func (nw *NetWork) DeleteNetWork(configPath string) error {
	if has, err := container.DirOrFileExist(path.Join(configPath, nw.Name)); err == nil && !has {
		log.Mylog.Error("DeleteNetWork", "container.DirOrFileExist", err)
		return err
	} else if err == nil && has {
		//说明文件存在
		return os.Remove(path.Join(configPath, nw.Name))
	}
	return nil
}

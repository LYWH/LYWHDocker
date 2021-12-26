package network

import (
	"LYWHDocker/log"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

/*
createTime:LYWH
createData:2021/12/25
*/

type BridgeNetWorkDriver struct {
}

//初始化网桥
func (b *BridgeNetWorkDriver) initBridge(n *NetWork) error {
	//1创建网络虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		log.Mylog.Error("create bridge interface failed", err)
		return err
	}
	//设置bridge的地址和路由
	gateWapIp := *n.IpRange
	gateWapIp.IP = n.IpRange.IP
	if err := setInterfaceIP(bridgeName, gateWapIp.String()); err != nil {
		log.Mylog.Error("error at assising ip", gateWapIp.String(), "on", bridgeName)
		return err
	}
	//启动bridge
	if err := setInterfaceUp(bridgeName); err != nil {
		log.Mylog.Error(err)
		return err
	}
	if err := setUpIPTables(bridgeName, n.IpRange); err != nil {
		log.Mylog.Error(err)
		return err
	}
	return nil
}

//创建bridge网络驱动
func createBridgeInterface(bridgeName string) error {
	//检验是否有同名的网络设备
	netInterface, err := net.InterfaceByName(bridgeName)
	if netInterface != nil || err == nil {
		log.Mylog.Error("createBridgeInterface", bridgeName, "exsited", err)
		return err
	}
	//创建link基础对象,其名字使用bridgeName
	nl := netlink.NewLinkAttrs()
	nl.Name = bridgeName
	//使用link创建netlink的bridge
	br := &netlink.Bridge{LinkAttrs: nl}
	//创建虚拟网络设备
	if err = netlink.LinkAdd(br); err != nil {
		log.Mylog.Error("createBridgeInterface", "netlink.LinkAdd", err)
		return err
	}
	return nil
}

//为创建的bridge分配IP和路由
func setInterfaceIP(name, IP string) error {
	//找到刚创立的bridge
	netInterface, err := netlink.LinkByName(name)
	if err != nil {
		log.Mylog.Error("err bridge name", err)
		return err
	}
	//netlink.ParseIPNet是对net.ParseCIDR的封装，返回的值ipNet中既包含了网段的信息(192.168.0.0/24)也包含了原始的IP地址(192.168.0.1)
	ipNet, err := netlink.ParseIPNet(IP)
	if err != nil {
		log.Mylog.Error(err)
		return err
	}
	// 通过netlink.AddrAdd给网络接口配置地址，等价于 ip addr add xxxx命令
	// 同时如果配置了地址所在的网段信息，例如192.168.0.0/24, 还会配置路由表192.168.0.0/24转发到这个bridge上
	addr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
		Flags: 0,
		Scope: 0,
	}
	//将addr信息加到netInterface中
	return netlink.AddrAdd(netInterface, addr)
}

//启动bridge
func setInterfaceUp(bridgeName string) error {
	//查找bridge
	bridgeInterface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		log.Mylog.Error("err bridge name", err)
		return err
	}
	if err = netlink.LinkSetUp(bridgeInterface); err != nil {
		log.Mylog.Error("setInterfaceUp", "netlink.LinkSetUp", err)
		return err
	}
	return nil
}

//设置iptable对应bridge的MASQUERADE规则
func setUpIPTables(bridgeName string, subnet *net.IPNet) error {
	iptableCMArgs := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptableCMArgs, " ")...)
	output, err := cmd.Output()
	if err != nil {
		log.Mylog.Error("iptables out", output)
		return err
	}
	return nil
}

func (b *BridgeNetWorkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetWorkDriver) Create(name, subnet string) (*NetWork, error) {
	//获取子网字符串的网关ip和网络段IP
	ip, ipRange, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Mylog.Error("Create", "net.ParseCIDR", err)
		return nil, err
	}
	ipRange.IP = ip
	//初始化网络对象
	n := &NetWork{
		Name:    name,
		IpRange: ipRange,
		Driver:  b.Name(),
	}
	//初始化网络配置
	if err = b.initBridge(n); err != nil {
		log.Mylog.Error("initBridge", "Create", err)
		return nil, err
	}
	return n, nil
}

//删除bridge网络
func (b *BridgeNetWorkDriver) Delete(netWork *NetWork) error {
	//根据名字查找网关
	bridgeInterface, err := netlink.LinkByName(netWork.Name)
	if err != nil {
		log.Mylog.Error("Delete", "netlink.LinkByName", err)
		return err
	}
	return netlink.LinkDel(bridgeInterface)
}

//创建veth，并且连接网络与veth端点
func (b *BridgeNetWorkDriver) Connect(netWork *NetWork, endPoint *EndPoint) error {
	//根据名字查找网关
	bridgeInterface, err := netlink.LinkByName(netWork.Name)
	if err != nil {
		log.Mylog.Error("Delete", "netlink.LinkByName", err)
		return err
	}
	//创建Veth接口配置
	la := netlink.NewLinkAttrs()
	//linux接口名的限制，取名字的前5位
	la.Name = endPoint.ID[:5]
	//设置veth的master接口属性，将veth的一端挂载到bridge中
	la.MasterIndex = bridgeInterface.Attrs().Index

	//创建veth对象，通过PeerName配置Veth的宁外一端，
	endPoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endPoint.ID[:5],
	}
	//调用netlink的LinkAdd方法创建Veth接口
	if err = netlink.LinkAdd(&endPoint.Device); err != nil {
		log.Mylog.Error("netlink.LinkAdd", "Connect", err)
		return err
	}
	//启动veth
	if err = netlink.LinkSetUp(&endPoint.Device); err != nil {
		log.Mylog.Error("netlink.LinkSetUp", "Connect", err)
		return err
	}
	return nil
}

//取消bridge桥接
func (b *BridgeNetWorkDriver) DisConnect(netWork *NetWork, endPoint *EndPoint) error {
	return nil
}

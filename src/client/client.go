package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
)

// see https://gitee.com/sjviper_admin/intranet-penetration/tree/master

type clientNg struct {
	ChanelMsg chan []byte
	Conn      net.Conn
	Target    net.Conn
	Name      string
}

func (c *clientNg) Read(ctx context.Context) {
	for {
		data := make([]byte, 4096)
		select {
		case <-ctx.Done():
			return
		default:
			n, err := c.Conn.Read(data)
			if err != nil {
				client.ExitMsg <- true
				return
			}
			c.ChanelMsg <- data[:n]
		}
	}
}

type Client struct {
	Intranet *clientNg
	Extranet *clientNg
	ExitMsg  chan bool
}

var client *Client

func initClientNg(host, port string) *clientNg {
	cg := new(clientNg)
	cg.ChanelMsg = make(chan []byte)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		panic(err.Error())
	}
	cg.Conn = conn
	if host != "" {
		fmt.Println("远程服务器连接成功\n")
	} else {
		fmt.Println("本地服务器连接成功\n")
	}

	return cg
}

//func initClient(localHost, remoteHost, inPort, exPort string) *Client {
//	client := new(Client)
//	client.ChanelMsg = make(chan []byte)
//	client.Intranet = initClientNg(localHost, inPort)
//	client.Extranet = initClientNg(remoteHost, exPort)
//	client.Intranet.Name = "内网客户端"
//	client.Extranet.Name = "外网客户端"
//	client.Intranet.Target = client.Extranet.Conn
//	client.Extranet.Target = client.Intranet.Conn
//	return client
//}

//func (cg *clientNg) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
//	err := cg.Target.AsyncWrite(packet)
//	if err != nil {
//		fmt.Println(cg.Name, "数据发送错误", err.Error())
//		_ = c.SendTo([]byte("服务端转发异常！"))
//	}
//	cg.Target.Wake()
//	return
//}

func NewClient(localHost, remoteHost, inPort, exPort string) *Client {
	client := new(Client)
	client.Intranet = initClientNg(localHost, inPort)
	client.Extranet = initClientNg(remoteHost, exPort)
	client.Intranet.Name = "内网客户端"
	client.Extranet.Name = "外网客户端"
	client.Intranet.Target = client.Extranet.Conn
	client.Extranet.Target = client.Intranet.Conn
	return client
}

func (c *Client) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	go c.Intranet.Read(ctx)
	go c.Extranet.Read(ctx)

	select {
	case data := <-c.Intranet.ChanelMsg:
		c.Extranet.Conn.Write(data)
	case data := <-c.Extranet.ChanelMsg:
		c.Intranet.Conn.Write(data)
	case <-c.ExitMsg:
		fmt.Println("程序退出")
		cancel()
		os.Exit(0)
	}
}

func (cg *clientNg) handle(packet []byte, c net.Conn) {
	for {
		_, err := cg.Target.Write(packet)
		if err != nil {
			fmt.Println(cg.Name, "数据发送错误", err.Error())
			c.Write([]byte("服务端转发异常！"))
		}
	}

}

func main() {
	var port string
	var remote string
	flag.StringVar(&port, "p", "9700", "需要穿透的端口")
	flag.StringVar(&remote, "h", "0.0.0.0", "远程服务地址")
	flag.Parse()
	fmt.Println("[服务启动成功][ok]|本地映射端口:", port, "｜外网访问地址：", remote, ":9001\n")
	client = NewClient("", remote, port, "9000")
	client.Run()
}

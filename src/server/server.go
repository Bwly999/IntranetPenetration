package main

import (
	"flag"
	"fmt"
	"github.com/panjf2000/gnet"
	"os"
	"strconv"
)

type echoServer struct {
	*gnet.EventServer
	clientMap map[string]gnet.Conn
	Name      string
}

var (
	InMap   = make(map[string]gnet.Conn)
	ExMap   = make(map[string]gnet.Conn)
	IN      = "in"
	OUT     = "out"
	exitMsg = make(chan bool)
	port    int
)

func initEchoServer(port, name string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			exitMsg <- true
		}
	}()
	sev := new(echoServer)
	sev.clientMap = make(map[string]gnet.Conn)
	sev.Name = name

	err := gnet.Serve(sev, "tcp://0.0.0.0:"+port, gnet.WithMulticore(true))
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(name, "服务启动成功 port:", port)
}

func initSNgrok(InPort, ExPort string) {
	go initEchoServer(InPort, IN)
	go initEchoServer(ExPort, OUT)
}

func (es *echoServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	if es.Name == IN {
		//fmt.Println(IN, "接受数据\n")
		for key, c := range ExMap {
			err := c.AsyncWrite(frame)
			if err != nil {
				fmt.Println(key, " ", OUT, "发送失败", err.Error())
			}
		}
	}

	if es.Name == OUT {
		//fmt.Println("接受数据\n")
		for key, c := range InMap {
			err := c.AsyncWrite(frame)
			if err != nil {
				fmt.Println(key, " ", IN, "发送失败", err.Error())
			}
		}
	}
	return
}

func (es *echoServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	fmt.Println(es.Name, "有一个新连接 [remoteIp:", c.RemoteAddr(), "]")
	if es.Name == IN {
		InMap[c.RemoteAddr().String()] = c
		return
	}
	ExMap[c.RemoteAddr().String()] = c
	return
}

func (es *echoServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	if es.Name == IN {
		fmt.Println(es.Name, " in连接已断开 [remoteIp:", c.RemoteAddr(), "]")
		if err != nil {
			fmt.Println(err.Error())
		}
		delete(InMap, c.RemoteAddr().String())
		return
	}
	fmt.Println(es.Name, " out连接已断开 [remoteIp:", c.RemoteAddr(), "]")
	if err != nil {
		fmt.Println(err.Error())
	}
	delete(ExMap, c.RemoteAddr().String())
	return
}

func init() {
	flag.IntVar(&port, "p", 39999, "穿透代理对外访问端口")
	flag.Parse()
}

func main() {
	initSNgrok("9000", strconv.Itoa(port))
	fmt.Println("[服务启动成功][ok][服务端口:9000,穿透外网访问端口:", port, "]")
	select {
	case <-exitMsg:
		fmt.Println("服务停止")
		os.Exit(0)
	}
}

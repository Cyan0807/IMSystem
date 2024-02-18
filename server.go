package main

import (
	"fmt"
	"net"
	"sync"
)

// 服务器端类
type Server struct {
	IP        string
	Port      int
	OnlineMap map[string]*User // 在线用户列表
	mapLock   sync.RWMutex
	Message   chan string // 消息广播的channel
}

// 启动一个服务端
func (this *Server) Start() {
	// 开始监听端口
	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", this.IP, this.Port))
	// 构建监听器时失败
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}
	//结束后关闭
	defer listener.Close()

	go this.ListenMessage()

	for true {
		// Accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("net.Accept err: ", err)
			continue
		}

		// Handler
		go this.Handle(conn)
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg + "\n"
	this.Message <- sendMsg
}

func (this *Server) Handle(conn net.Conn) {
	//fmt.Println("连接建立成功")
	// 用户上线
	user := NewUser(conn)
	// 加入onlineMap
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()
	// 广播用户上线的消息
	this.BroadCast(user, "已上线")

	// 当前handler阻塞
	select {}
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线User
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		// 将信息发送给所有在线用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// ====================================================

// 创建一个服务器端
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

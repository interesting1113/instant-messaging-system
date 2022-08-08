package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 监听Message广播消息的goroutine, 一旦有消息就发送给全部在线的user
func (this *Server) ListenMessage()  {
	for {
		msg := <- this.Message

		// 将消息发送给全部在线的user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}

	return server
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn)  {
	// ...当前链接业务
	fmt.Println("链接建立成功")
	user := NewUser(conn, this)

	// 用户上线， 将用户加入到OnlineMap中
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err: ", err)
				return
			}

			msg := string(buf[:n - 1])

			// 将得到的消息进行广播
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是活跃的
			isLive <- true
		}
	}()
	// 当前handler阻塞
	for {
		select {
			case <- isLive:
				// 当前用户是活跃的，应该重置定时器
				// 不做任何处理，为了激活select，更新下面的定时器

			case <- time.After(time.Second * 10):
			// 已经超时
			// 将当前的客户端User强制关闭
			user.sendMsg("你被踢了")

			// 销毁用的资源
			close(user.C)

			// 关闭链接
			conn.Close()

			// 推出当前的handler
			return // runtime.Goexit()
		}
	}

}

func (this *Server) Start()  {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Print("new Listen error:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}
}
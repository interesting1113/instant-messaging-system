package main

import (
	"net"
)

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn

	server *Server // 当前用户属于哪个server
}

// 创建一个用户的api
func NewUser(conn net.Conn, server *Server)  *User{
	userAddr := conn.RemoteAddr().String()

	user := &User {
		Name: userAddr,
		Addr: userAddr,
		C: make(chan  string),
		conn: conn,
		server: server,
	}
	// 启动监听当前user channel的构成
	go user.ListenMessage()

	return user
}

// 用户的上线业务
func (this *User) Online()  {
	// 用户上线， 将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")

}

// 用户的下线业务
func (this *User) Offline()  {
	// 用户上线， 将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已下线")
}

// 给当前user对应的客户端发送消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息业务
func (this *User) DoMessage(msg string)  {
	if msg == "who" {
		// 查询都有哪些用户在线
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线...\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else {
		this.server.BroadCast(this, msg)
	}
}

func (this *User) ListenMessage()  {
	for {
		msg := <- this.C
		this.conn.Write([]byte(msg +"\n"))
	}
}

package main

import (
	"net"
	"strings"
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
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询都有哪些用户在线
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式rename|
		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名已被使用")
		} else if len(msg) > 4 && msg[:3] == "to|" {
			// 消息格式to|
			// 1. 获取对方的用户名
			remoteName := strings.Split(msg, "|")[1]
			if remoteName == "" {
				this.SendMsg("消息格式不正确， 请使用\"to|hyde|hello\"格式。\n")
				return
			}
			// 2. 根据用户名得到对方的user对象
			remoteUser, ok := this.server.OnlineMap[remoteName]
			if !ok {
				this.SendMsg("该用户名不存在")
				return
			}
			// 3. 获取消息内容，通过对方的User对象将消息发送过去
			content := strings.Split(msg, "|")[2]
			if content == "" {
				this.SendMsg("无消息内容，请重新发送")
				return
			}
			remoteUser.SendMsg(this.Name + "对您说：" + content)
		}else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已更新用户名" + this.Name + "\n")
		}

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

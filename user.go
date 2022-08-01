package main

import (
	"net"
)

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn
}

// 创建一个用户的api
func NewUser(conn net.Conn)  *User{
	userAddr := conn.RemoteAddr().String()

	user := &User {
		Name: userAddr,
		Addr: userAddr,
		C: make(chan  string),
		conn: conn,
	}
	// 启动监听当前user channel的构成
	go user.ListenMessage()

	return user
}

func (this *User) ListenMessage()  {
	for {
		msg := <- this.C
		this.conn.Write([]byte(msg +"\n"))
	}
}

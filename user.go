package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser create a user API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// launch the goroutine of listening to current user channel
	go user.ListenMessage()

	return user
}

func (this *User) Online() {
	// when user online, add it into onlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// tell other users a user is online
	this.server.BoardCast(this, "online")
}

func (this *User) Offline() {
	// when user offline, delete from onlineMap
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// tell other users a user is offline
	this.server.BoardCast(this, "offline")
}

func (this *User) DoMessage(msg string) {
	this.server.BoardCast(this, msg)
}

// ListenMessage listen to current channel, once have message, send it to user client
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

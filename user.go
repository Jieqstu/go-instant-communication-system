package main

import (
	"net"
	"strings"
)

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

// SendMsg send msg to corresponding user client
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// search all online users
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "is online...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// format: rename|newName
		newName := strings.Split(msg, "|")[1]
		//check whether newName is used
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("username is used\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("username is updated:" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// format: to|remoteName|content
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("wrong format, please use \"to|remoteName|content\".\n")
			return
		}

		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("username not found")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("empty content\n")
			return
		}

		remoteUser.SendMsg(this.Name + " says: " + content)
	} else {
		this.server.BoardCast(this, msg)
	}
}

// ListenMessage listen to current channel, once have message, send it to user client
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

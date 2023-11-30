package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		// send message to all online users
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BoardCast method for telling other users a user is online
func (this *Server) BoardCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("connect successfully")
	user := NewUser(conn, this)

	// when user online
	user.Online()

	// a channel to listen to whether current user is active
	isActive := make(chan bool)

	//accept message send from client
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// extract user's message and eliminate '\n' at the end
			msg := string(buf[:n-1])

			user.DoMessage(msg)

			// user is active if it has any message
			isActive <- true
		}
	}()

	// block handler
	for {
		select {
		case <-isActive:
			// user is active, reset timer
			// do noting, so we can use select update the timer
		case <-time.After(time.Second * 300):
			// timeout
			user.SendMsg("kicked\n")

			// close user's channel
			close(user.C)

			// close connection
			conn.Close()

			return
		}
	}
}

func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	//close listen socket
	defer listener.Close()

	//start goroutine of listening Message
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}

}

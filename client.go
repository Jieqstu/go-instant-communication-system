package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       99,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}

	client.conn = conn

	return client
}

// DealResponse deal with message coming from server, standard output
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.Boardcast")
	fmt.Println("2.Private Chat")
	fmt.Println("3.Update Username")
	fmt.Println("0.Exit")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>input number out of bounds<<<")
		return false
	}
}

func (client *Client) SelectOnlineUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error: ", err)
		return
	}
}

func (client *Client) PrivateChar() {
	var (
		remoteName string
		chatMsg    string
	)
	client.SelectOnlineUser()
	fmt.Println(">>> Input username you want to talk, or exit: ")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println(">>> Input message content, or exit: ")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error: ", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>> Input message content, or exit: ")
			fmt.Scanln(&chatMsg)
		}

		client.SelectOnlineUser()
		fmt.Println(">>> Input username you want to talk, or exit: ")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) Boardcast() {
	var boardcastMsg string

	fmt.Println(">>> Input boardcast content, or exit: ")
	fmt.Scanln(&boardcastMsg)

	for boardcastMsg != "exit" {
		if len(boardcastMsg) != 0 {
			sendMsg := boardcastMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error: ", err)
				break
			}
		}

		boardcastMsg = ""
		fmt.Println(">>> Input boardcast content, or exit")
		fmt.Scanln(&boardcastMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>> Input new username: ")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}

		switch client.flag {
		case 1:
			client.Boardcast()
			break
		case 2:
			client.PrivateChar()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

var (
	serverIp   string
	serverPort int
)

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set server IP(default: 127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "set server Port(default: 8888)")
}

func main() {
	// parse command line
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> fail to connect to servery...")
		return
	}
	// have a separate goroutine to handle message coming from server
	go client.DealResponse()
	fmt.Println(">>>>> connect to server successfully...")

	client.Run()
}

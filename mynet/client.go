package mynet

import (
	"onigoko/data"
)

type Communicator interface {
	//向服务端发送消息
	sendMessage(operation data.Operation) error

	//持续接受消息并且通过chan传递给render
	receiveMessage() error
}

//Client 负责前后端通信，发送指令，接收指令
type Client struct {
}

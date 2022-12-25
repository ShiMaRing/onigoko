package mynet

import "onigoko/data"

const DefaultAddr = "ws://127.0.0.1"
const DefaultPort = 8085

type Communicator interface {
	// SendMessage 向服务端发送消息
	SendMessage(operation data.Operation)

	sendMessage() error

	// ReceiveMessage 持续接受消息并且通过chan传递给render
	receiveMessage() error

	// Run 持续发送消息以及接收消息
	Run() error

	ReceivedMessage() chan data.Operation
}

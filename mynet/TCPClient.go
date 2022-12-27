package mynet

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"onigoko/data"
	"sync"
)

//可能更换地址，便于测试
var serverAddress = []string{"127.0.0.1:8888", "127.0.0.1:8889", "127.0.0.1:8890"}

// TCPClient  tcp client
type TCPClient struct {
	toSend       chan data.Operation
	receivedChan chan data.Operation
	conn         net.Conn
	bufWriter    *bufio.Writer
	mu           sync.Mutex // 互斥锁
}

func NewTCPClient() *TCPClient {
	return &TCPClient{
		receivedChan: make(chan data.Operation, 10),
		toSend:       make(chan data.Operation, 10),
	}
}

func (t *TCPClient) SendMessage(operation data.Operation) {
	t.toSend <- operation
}

func (t *TCPClient) sendMessage() error {
	operation := <-t.toSend
	bytes, err := json.Marshal(operation)
	if err != nil {
		log.Fatalln("Marshal", err)
		return err
	}
	println("发送消息：", string(bytes))
	println("发送消息大小：", len(bytes))
	t.conn.Write(bytes)
	return err
}

func (t *TCPClient) ReceiveMessage() error {
	//接收消息
	//log here
	var err error
	//首先读取消息头
	header := make([]byte, 4)
	_, _ = t.conn.Read(header)
	messageSize := binary.BigEndian.Uint32(header)
	fmt.Println("接收消息大小：", messageSize)
	//根据消息头的大小，读取消息体
	message := make([]byte, messageSize)
	//循环读取，直到读取到足够的字节
	var count int
	for count < int(messageSize) {
		n, err := t.conn.Read(message[count:])
		if err != nil && !errors.As(err, &io.EOF) {
			return err
		}
		count += n
	}
	//log here
	fmt.Println("接收到的消息：", string(message))
	var op data.Operation
	err = json.Unmarshal(message, &op)
	if err != nil {
		log.Fatalln("Unmarshal", err)
	}
	t.receivedChan <- op
	return nil
}

// Run 尝试连接服务器
func (t *TCPClient) Run() error {
	var err error
	for i := range serverAddress {
		t.conn, err = net.Dial("tcp", serverAddress[i])
		if err != nil {
			fmt.Println("连接服务器失败，正在尝试连接下一个服务器,错误信息：", err)
			continue
		} else {
			break
		}
	}
	if t.conn == nil {
		log.Fatalln("连接服务器失败", err)
		return err
	}
	t.bufWriter = bufio.NewWriter(t.conn)
	go func() {
		defer func() {
			close(t.receivedChan)
			t.conn.Close()
		}()
		for true {
			if revErr := t.ReceiveMessage(); revErr != nil {
				log.Fatalln("receiveMessage:", err)
				return
			}
		}
	}()
	go func() {
		defer func() {
			close(t.toSend)
		}()
		for true {
			if sendErr := t.sendMessage(); sendErr != nil {
				log.Fatalln("sendMessage:", err)
				return
			}
		}
	}()
	return nil
}

func (t *TCPClient) ReceivedMessage() chan data.Operation {
	return t.receivedChan
}

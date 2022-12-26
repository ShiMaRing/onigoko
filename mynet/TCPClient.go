package mynet

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	fmt.Println("发送消息：", operation)
	bytes, err := json.Marshal(operation)
	if err != nil {
		log.Fatalln("Marshal", err)
		return err
	}
	var count int
	for count < len(bytes) {
		nn, err := t.bufWriter.Write(bytes)
		if err != nil {
			return err
		}
		count += nn
	}
	err = t.bufWriter.Flush()
	return err
}

func (t *TCPClient) ReceiveMessage() error {
	//接收消息
	//log here
	var err error
	buf := make([]byte, 1024)
	result := make([]byte, 0)
	var count = 0
	for true {
		n, _ := t.conn.Read(buf)
		result = append(result, buf[:n]...)
		count += n
		if n < 1024 {
			break
		}
	}
	var op data.Operation
	err = json.Unmarshal(result[:count], &op)

	tmp := op
	tmp.Blocks = nil
	marshal, _ := json.Marshal(tmp)
	fmt.Println("接收消息：", string(marshal))

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

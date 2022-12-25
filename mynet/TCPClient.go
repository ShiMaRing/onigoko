package mynet

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/netip"
	"onigoko/data"
)

//可能更换地址，便于测试
var serverAddress = []string{"127.0.0.1:8088", "127.0.0.1:8089", "127.0.0.1:8090"}

// TCPClient  tcp client
type TCPClient struct {
	toSend       chan data.Operation
	receivedChan chan data.Operation
	conn         *net.TCPConn
	bufReader    *bufio.Reader
	bufWriter    *bufio.Writer
}

func NewTCPClient() *TCPClient {
	return &TCPClient{
		receivedChan: make(chan data.Operation),
		toSend:       make(chan data.Operation),
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

func (t *TCPClient) receiveMessage() error {
	//接收消息
	buf := make([]byte, 4096)
	n, err := t.bufReader.Read(buf)
	if err != nil {
		return err
	}
	var op data.Operation
	err = json.Unmarshal(buf[:n], &op)
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
		addr, err := netip.ParseAddrPort(serverAddress[i])
		if err != nil {
			panic(err)
		}
		t.conn, err = net.DialTCP("tcp", net.TCPAddrFromAddrPort(addr), nil)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if t.conn == nil {
		return err
	}
	t.bufReader = bufio.NewReader(t.conn)
	t.bufWriter = bufio.NewWriter(t.conn)
	go func() {
		defer func() {
			close(t.receivedChan)
			t.conn.Close()
		}()
		for true {
			if revErr := t.receiveMessage(); revErr != nil {
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

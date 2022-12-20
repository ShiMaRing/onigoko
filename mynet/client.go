package mynet

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"onigoko/data"
)

//Client 负责前后端通信，发送指令，接收指令,将接收到的指令添加到chan中，客户端读取后更新游戏状态
type Client struct {
	toSend     chan data.Operation
	received   chan data.Operation
	conn       *websocket.Conn //前后端websocket通信
	serverIp   string
	serverPort int
}

func (c *Client) ReceivedMessage() chan data.Operation {
	return c.received
}

// NewDefaultClient 默认客户端，使用默认服务器地址
func NewDefaultClient() *Client {
	return &Client{
		toSend:     make(chan data.Operation, 10),
		received:   make(chan data.Operation, 10),
		conn:       nil,
		serverIp:   DefaultAddr,
		serverPort: DefaultPort,
	}
}

func (c *Client) dial() error {
	dialer := websocket.Dialer{}
	//向服务器发送连接请求，websocket 统一使用 ws://，默认端口和http一样都是80
	conn, _, err := dialer.Dial(fmt.Sprintf("%s:%d", c.serverIp, c.serverPort), nil)
	if nil != err {
		return err
	}
	c.conn = conn
	return nil
}

//发送消息
func (c *Client) SendMessage(operation data.Operation) {
	c.toSend <- operation
	return
}

func (c *Client) receiveMessage() error {
	defer c.conn.Close()
	var err error
	for {
		op := data.Operation{}
		messageType, p, innerErr := c.conn.ReadMessage()
		if messageType != websocket.TextMessage {
			continue //不处理文本消息以外的消息
		}
		if innerErr != nil {
			return err
		}
		err = json.Unmarshal(p, &op)
		if err != nil {
			log.Fatalln("json unmarshal fail:", err)
		}
		c.received <- op //将消息置入消费池
	}
	return nil
}

func (c *Client) Run() error {
	err := c.dial()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case op := <-c.toSend:
				jsonData, _ := json.Marshal(op)
				err := c.conn.WriteMessage(websocket.TextMessage, jsonData)
				if err != nil {
					log.Fatalln(err)
					return
				}
			}
		}
	}()
	if err := c.receiveMessage(); err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}

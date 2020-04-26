package chat

import (
	"fmt"
	"io"

	"golang.org/x/net/websocket"
)

const channelBufSize = 100

var maxId int = 0

type Client struct {
	id     int
	ws     *websocket.Conn
	server *Server
	ch     chan *Message
	doneCh chan bool
}

func NewClient(ws *websocket.Conn, server *Server) *Client {
	if ws == nil {
		panic("ws cannot be nil")
	}
	if server == nil {
		panic("server canno be nil")
	}
	maxId++
	ch := make(chan *Message, channelBufSize)
	doneCh := make(chan bool)
	return &Client{maxId, ws, server, ch, doneCh}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Write(msg *Message) {
	select {
	case c.ch <- msg:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected", c.id)
		c.server.Error(err)
	}
}

func (c *Client) Done() {
	c.doneCh <- true
}

func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

func (c *Client) listenWrite() {
	// Бесконечный цикл на прослушку записи.
	// Как только в канал c.ch приходил сообщение,
	// отправляем его в сокет. Если прилетает в канал c.doneCh,
	// завершаем цикл.
	for {
		select {
		// Send message to the client
		case msg := <-c.ch:
			websocket.JSON.Send(c.ws, msg)
		// Receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true
			return
		}
	}
}

func (c *Client) listenRead() {
	// Бесконечный цикл на прослушку каналов.
	// По приходу сообщения в канал c.doneCh - завершает цикл
	// По умолчанию опрашивает сокет и как только появляетяс сообщение,
	// отдает серверу для массовой рассылки всем клиентам.
	for {
		select {
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true
			return
		default:
			var msg Message
			err := websocket.JSON.Receive(c.ws, &msg)
			if err == io.EOF {
				c.doneCh <- true
			} else if err != nil {
				c.server.Error(err)
			} else {
				c.server.SendAll(&msg)
			}
		}
	}

}

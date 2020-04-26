package chat

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	pattern   string
	messages  []*Message
	clients   map[int]*Client
	addCh     chan *Client
	delCh     chan *Client
	sendAllCh chan *Message
	doneCh    chan bool
	errCh     chan error
}

func NewServer(pattern string) *Server {
	messages := []*Message{}
	clients := make(map[int]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	return &Server{
		pattern,
		messages,
		clients,
		addCh,
		delCh,
		sendAllCh,
		doneCh,
		errCh,
	}
}

func (s *Server) Add(c *Client) {
	s.addCh <- c
}

func (s *Server) Del(c *Client) {
	s.delCh <- c
}

func (s *Server) SendAll(msg *Message) {
	s.sendAllCh <- msg
}

func (s *Server) Done() {
	s.doneCh <- true
}

func (s *Server) Error(err error) {
	s.errCh <- err
}

func (s *Server) sendPastMessages(c *Client) {
	for _, msg := range s.messages {
		c.Write(msg)
	}
}

func (s *Server) sendAll(msg *Message) {
	for _, c := range s.clients {
		c.Write(msg)
	}
}

func (s *Server) Listen() {
	onConnected := func(ws *websocket.Conn) {
		// Создаем нового клиента,
		// добавляем его на сервер и запускаем прослушку
		defer func() {
			err := ws.Close()
			if err != nil {
				s.errCh <- err
			}
		}()

		client := NewClient(ws, s)
		s.Add(client)
		client.Listen()
	}
	http.Handle(s.pattern, websocket.Handler(onConnected))

	for {
		select {
		case c := <-s.addCh:
			s.clients[c.id] = c
			s.sendPastMessages(c)
		case c := <-s.delCh:
			delete(s.clients, c.id)
		case msg := <-s.sendAllCh:
			s.messages = append(s.messages, msg)
			s.sendAll(msg)
		case err := <-s.errCh:
			log.Println("Error:", err.Error())
		case <-s.doneCh:
			return
		}

	}

}

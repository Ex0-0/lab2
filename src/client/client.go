package client

import (
	"fmt"
	"net/http"
	"os"
	"lab2/src/tracer"
	"time"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

type Client struct {
	socket *websocket.Conn
	send   chan []byte
	room   *Room
	Name   string
	Token  string
}

var upgrader = &websocket.Upgrader{WriteBufferSize: socketBufferSize, ReadBufferSize: socketBufferSize}

func (c *Client) Read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			formattedMsg := fmt.Sprintf("%s: %s", c.Name, string(msg))
			c.room.Forward <- []byte(formattedMsg)
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *Client) Write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}

type Room struct {
	Forward chan []byte
	Join    chan *Client
	Leave   chan *Client
	Clients map[*Client]bool
	Tracer  tracer.Tracer
}

func NewRoom() *Room {
	return &Room{
		Forward: make(chan []byte),
		Join:    make(chan *Client),
		Leave:   make(chan *Client),
		Clients: make(map[*Client]bool),
	}
}

func generateToken(name string) string {
	return fmt.Sprintf("token_%d_%s", time.Now().UnixNano(), name)
}

func (r *Room) Run() {
	r.Tracer.Trace("🚀 Сервер чата запущен и готов к работе")
	r.Tracer.Trace("📡 Ожидание подключений...")
	r.Tracer.Trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	for {
		select {
		case client := <-r.Join:
			r.Clients[client] = true
			r.Tracer.Tracef("✅ ПОДКЛЮЧЕНИЕ | Имя: %-15s | Токен: %s", client.Name, client.Token)
			r.Tracer.Tracef("👥 Всего клиентов в чате: %d", len(r.Clients))
			
		case client := <-r.Leave:
			delete(r.Clients, client)
			close(client.send)
			r.Tracer.Tracef("❌ ОТКЛЮЧЕНИЕ | Имя: %-15s | Причина: выход из чата", client.Name)
			r.Tracer.Tracef("👥 Всего клиентов в чате: %d", len(r.Clients))
			
		case msg := <-r.Forward:
			r.Tracer.Tracef("💬 СООБЩЕНИЕ | %s", string(msg))
			sentCount := 0
			failedCount := 0
			
			for client := range r.Clients {
				select {
				case client.send <- msg:
					sentCount++
				default:
					delete(r.Clients, client)
					close(client.send)
					failedCount++
				}
			}
			
			if sentCount > 0 {
				r.Tracer.Tracef("📤 Доставлено сообщений: %d", sentCount)
			}
			if failedCount > 0 {
				r.Tracer.Tracef("⚠️  Не доставлено (клиент отключен): %d", failedCount)
			}
			r.Tracer.Trace("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄")
		}
	}
}

func (r *Room) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка WebSocket: %v", err)
		return
	}

	name := req.URL.Query().Get("name")
	if name == "" {
		name = "Аноним"
	}

	token := req.URL.Query().Get("token")
	if token == "" {
		token = generateToken(name)
	}

	client := &Client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
		Name:   name,
		Token:  token,
	}

	r.Join <- client

	client.socket.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("=== Ваш токен: %s ===", token)))

	defer func() {
		r.Leave <- client
	}()

	go client.Write()
	client.Read()
}

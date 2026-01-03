package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	ID     uuid.UUID
	Conn   *ws.Conn
	Send   chan []byte
	UserID uuid.UUID
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Hub struct {
	Clients    map[uuid.UUID]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.UserID)
			h.broadcastStatus(client.UserID, true)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
				h.mu.Unlock()
				log.Printf("Client unregistered: %s", client.UserID)
				h.broadcastStatus(client.UserID, false)
			} else {
				h.mu.Unlock()
			}

		case message := <-h.Broadcast:
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("error marshaling message: %v", err)
				continue
			}
			h.mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(h.Clients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcastStatus(userID uuid.UUID, online bool) {
	msg := Message{
		Type: "user_status",
		Content: map[string]interface{}{
			"user_id": userID,
			"online":  online,
		},
	}
	h.Broadcast <- msg
}

// Проверить, онлайн ли пользователь
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.Clients[userID]
	return ok
}

// Отправить сообщение конкретному пользователю
func (h *Hub) SendToUser(userID uuid.UUID, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}

	h.mu.RLock()
	client, ok := h.Clients[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			h.mu.Lock()
			delete(h.Clients, userID)
			h.mu.Unlock()
		}
	}
}

func (c *Client) ReadPump(h *Hub) {
	defer func() {
		h.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

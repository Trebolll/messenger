package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

// WallClient — участник комнаты публичного чата стены
type WallClient struct {
	ID     uuid.UUID
	UserID *uuid.UUID // nil если не авторизован (только чтение)
	Conn   *ws.Conn
	Send   chan []byte
	ChatID uuid.UUID
}

// WallHub — хаб с комнатами, каждая комната = chat_id поста
type WallHub struct {
	// rooms: chat_id -> set of clients
	rooms      map[uuid.UUID]map[*WallClient]struct{}
	register   chan *WallClient
	unregister chan *WallClient
	broadcast  chan wallBroadcast
	mu         sync.RWMutex
}

type wallBroadcast struct {
	chatID  uuid.UUID
	payload []byte
}

func NewWallHub() *WallHub {
	return &WallHub{
		rooms:      make(map[uuid.UUID]map[*WallClient]struct{}),
		register:   make(chan *WallClient, 256),
		unregister: make(chan *WallClient, 256),
		broadcast:  make(chan wallBroadcast, 512),
	}
}

func (h *WallHub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.rooms[c.ChatID] == nil {
				h.rooms[c.ChatID] = make(map[*WallClient]struct{})
			}
			h.rooms[c.ChatID][c] = struct{}{}
			h.mu.Unlock()
			log.Printf("WallClient registered for room %s", c.ChatID)

		case c := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[c.ChatID]; ok {
				if _, ok := room[c]; ok {
					delete(room, c)
					close(c.Send)
					if len(room) == 0 {
						delete(h.rooms, c.ChatID)
					}
				}
			}
			h.mu.Unlock()

		case b := <-h.broadcast:
			h.mu.RLock()
			room, ok := h.rooms[b.chatID]
			if !ok {
				h.mu.RUnlock()
				log.Printf("WallHub: No room %s to broadcast", b.chatID)
				continue
			}
			clients := make([]*WallClient, 0, len(room))
			for c := range room {
				clients = append(clients, c)
			}
			h.mu.RUnlock()

			log.Printf("WallHub: Broadcasting to %d clients in room %s", len(clients), b.chatID)
			for _, c := range clients {
				select {
				case c.Send <- b.payload:
				default:
					h.mu.Lock()
					if room, ok := h.rooms[c.ChatID]; ok {
						if _, ok := room[c]; ok {
							delete(room, c)
							close(c.Send)
							if len(room) == 0 {
								delete(h.rooms, c.ChatID)
							}
						}
					}
					h.mu.Unlock()
				}
			}
		}
	}
}

func (h *WallHub) Register() chan<- *WallClient {
	return h.register
}

// BroadcastToRoom — отправить сообщение всем в комнате (чат поста или стена пользователя)
func (h *WallHub) BroadcastToRoom(roomID uuid.UUID, msg interface{}) {
	log.Printf("WallHub: Broadcasting to room %s", roomID)
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WallHub marshal error: %v", err)
		return
	}
	h.broadcast <- wallBroadcast{chatID: roomID, payload: data}
}

func (c *WallClient) ReadPump(h *WallHub, onMessage func(c *WallClient, data []byte)) {
	defer func() {
		h.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("WallClient read error: %v", err)
			}
			break
		}
		if onMessage != nil {
			onMessage(c, data)
		}
	}
}

func (c *WallClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
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

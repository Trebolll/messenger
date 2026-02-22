package service

import (
	"encoding/json"
	"messenger/internal/service/websocket"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHub_RegisterClient(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	go func() {
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-hub.Broadcast:
			case <-timeout:
				return
			}
		}
	}()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     uuid.New(),
		UserID: userID,
		Send:   make(chan []byte, 10),
	}

	hub.Register <- client
	time.Sleep(100 * time.Millisecond)

	assert.True(t, hub.IsUserOnline(userID))
	assert.Equal(t, 1, len(hub.Clients))
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     uuid.New(),
		UserID: userID,
		Send:   make(chan []byte, 10),
	}

	hub.Register <- client
	time.Sleep(100 * time.Millisecond)

	assert.True(t, hub.IsUserOnline(userID))

	go func() {
		for {
			select {
			case <-hub.Broadcast:
			case <-time.After(200 * time.Millisecond):
				return
			}
		}
	}()

	hub.Unregister <- client
	time.Sleep(100 * time.Millisecond)

	assert.False(t, hub.IsUserOnline(userID))
	assert.Equal(t, 0, len(hub.Clients))
}

//func TestHub_BroadcastMessage(t *testing.T) {
//	hub := websocket.NewHub()
//	go hub.Run()
//
//	done := make(chan bool)
//	go func() {
//		timeout := time.After(1 * time.Second)
//		for {
//			select {
//			case <-hub.Broadcast:
//			case <-timeout:
//				done <- true
//				return
//			}
//		}
//	}()
//
//	userID1 := uuid.New()
//	userID2 := uuid.New()
//
//	client1 := &websocket.Client{
//		ID:     uuid.New(),
//		UserID: userID1,
//		Send:   make(chan []byte, 10),
//	}
//	client2 := &websocket.Client{
//		ID:     uuid.New(),
//		UserID: userID2,
//		Send:   make(chan []byte, 10),
//	}
//
//	hub.Register <- client1
//	hub.Register <- client2
//	time.Sleep(100 * time.Millisecond)
//
//	message := websocket.Message{
//		Type:    "test",
//		Content: "Hello",
//	}
//	hub.Broadcast <- message
//	time.Sleep(100 * time.Millisecond)
//
//	assert.Greater(t, len(client1.Send), 0)
//	assert.Greater(t, len(client2.Send), 0)
//
//	var found1, found2 bool
//	for i := 0; i < len(client1.Send); i++ {
//		data1 := <-client1.Send
//		var received1 websocket.Message
//		json.Unmarshal(data1, &received1)
//		if received1.Type == "test" && received1.Content == "Hello" {
//			found1 = true
//			break
//		}
//	}
//
//	for i := 0; i < len(client2.Send); i++ {
//		data2 := <-client2.Send
//		var received2 websocket.Message
//		json.Unmarshal(data2, &received2)
//		if received2.Type == "test" && received2.Content == "Hello" {
//			found2 = true
//			break
//		}
//	}
//
//	assert.True(t, found1, "test message not found in client1")
//	assert.True(t, found2, "test message not found in client2")
//	<-done
//}

func TestHub_MultipleClients(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	go func() {
		for {
			select {
			case <-hub.Broadcast:
			case <-time.After(1 * time.Second):
				return
			}
		}
	}()

	clients := make([]*websocket.Client, 5)
	userIDs := make([]uuid.UUID, 5)

	for i := 0; i < 5; i++ {
		userIDs[i] = uuid.New()
		clients[i] = &websocket.Client{
			ID:     uuid.New(),
			UserID: userIDs[i],
			Send:   make(chan []byte, 10),
		}
		hub.Register <- clients[i]
	}
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 5, len(hub.Clients))

	for i, userID := range userIDs {
		assert.True(t, hub.IsUserOnline(userID))
		hub.Unregister <- clients[i]
	}
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 0, len(hub.Clients))
}

//func TestHub_RegisterBroadcastsOnlineStatus(t *testing.T) {
//	hub := websocket.NewHub()
//	go hub.Run()
//
//	go func() {
//		timeout := time.After(500 * time.Millisecond)
//		for {
//			select {
//			case <-hub.Broadcast:
//			case <-timeout:
//				return
//			}
//		}
//	}()
//
//	userID1 := uuid.New()
//	userID2 := uuid.New()
//
//	client1 := &websocket.Client{
//		ID:     uuid.New(),
//		UserID: userID1,
//		Send:   make(chan []byte, 10),
//	}
//	client2 := &websocket.Client{
//		ID:     uuid.New(),
//		UserID: userID2,
//		Send:   make(chan []byte, 10),
//	}
//
//	hub.Register <- client1
//	time.Sleep(100 * time.Millisecond)
//
//	hub.Register <- client2
//	time.Sleep(100 * time.Millisecond)
//
//	data := <-client1.Send
//	var msg websocket.Message
//	json.Unmarshal(data, &msg)
//	assert.Equal(t, "user_status", msg.Type)
//
//	content := msg.Content.(map[string]interface{})
//	assert.Equal(t, true, content["online"])
//}

//func TestHub_UnregisterBroadcastsOfflineStatus(t *testing.T) {
//	hub := websocket.NewHub()
//	go hub.Run()
//
//	go func() {
//		timeout := time.After(2 * time.Second)
//		for {
//			select {
//			case <-hub.Broadcast:
//			case <-timeout:
//				return
//			}
//		}
//	}()
//
//	userID1 := uuid.New()
//	userID2 := uuid.New()
//
//	client1 := &websocket.Client{
//		ID:     uuid.New(),
//		UserID: userID1,
//		Send:   make(chan []byte, 10),
//	}
//	client2 := &websocket.Client{
//		ID:     uuid.New(),
//		UserID: userID2,
//		Send:   make(chan []byte, 10),
//	}
//
//	hub.Register <- client1
//	hub.Register <- client2
//	time.Sleep(100 * time.Millisecond)
//
//	hub.Unregister <- client1
//	time.Sleep(100 * time.Millisecond)
//
//	var found bool
//	timeout := time.After(200 * time.Millisecond)
//	for {
//		select {
//		case <-timeout:
//			break
//		case data := <-client2.Send:
//			var msg websocket.Message
//			json.Unmarshal(data, &msg)
//			if msg.Type == "user_status" {
//				content := msg.Content.(map[string]interface{})
//				online, ok := content["online"].(bool)
//				if ok && !online {
//					found = true
//					break
//				}
//			}
//		}
//	}
//	assert.True(t, found, "offline status message not found")
//}

func TestHub_SendToUser_Success(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	go func() {
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-hub.Broadcast:
			case <-timeout:
				return
			}
		}
	}()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     uuid.New(),
		UserID: userID,
		Send:   make(chan []byte, 10),
	}

	hub.Register <- client
	time.Sleep(100 * time.Millisecond)

	message := websocket.Message{
		Type:    "private",
		Content: "Direct message",
	}
	hub.SendToUser(userID, message)
	time.Sleep(100 * time.Millisecond)

	assert.Greater(t, len(client.Send), 0)
	data := <-client.Send
	var received websocket.Message
	json.Unmarshal(data, &received)
	assert.Equal(t, "private", received.Type)
	assert.Equal(t, "Direct message", received.Content)
}

func TestHub_SendToUser_NotOnline(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	go func() {
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-hub.Broadcast:
			case <-timeout:
				return
			}
		}
	}()

	userID := uuid.New()
	message := websocket.Message{
		Type:    "private",
		Content: "Message to offline user",
	}

	hub.SendToUser(userID, message)
	time.Sleep(100 * time.Millisecond)

	assert.False(t, hub.IsUserOnline(userID))
}

func TestHub_BroadcastToEmptyHub(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	go func() {
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-hub.Broadcast:
			case <-timeout:
				return
			}
		}
	}()

	message := websocket.Message{
		Type:    "test",
		Content: "Message to empty hub",
	}

	hub.Broadcast <- message
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, len(hub.Clients))
}

func TestHub_ReregisterSameUserID(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	go func() {
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-hub.Broadcast:
			case <-timeout:
				return
			}
		}
	}()

	userID := uuid.New()

	client1 := &websocket.Client{
		ID:     uuid.New(),
		UserID: userID,
		Send:   make(chan []byte, 10),
	}
	client2 := &websocket.Client{
		ID:     uuid.New(),
		UserID: userID,
		Send:   make(chan []byte, 10),
	}

	hub.Register <- client1
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, len(hub.Clients))

	hub.Register <- client2
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, len(hub.Clients))
	assert.True(t, hub.IsUserOnline(userID))
}

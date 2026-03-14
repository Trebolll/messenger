package service

import (
	"encoding/json"
	"messenger/internal/service/websocket"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWallHub_NewWallHub(t *testing.T) {
	hub := websocket.NewWallHub()
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.Register())
	assert.NotNil(t, hub.Unregister())
}

func TestWallHub_Register(t *testing.T) {
	hub := websocket.NewWallHub()
	go hub.Run()

	chatID := uuid.New()
	client := &websocket.WallClient{
		ID:     uuid.New(),
		ChatID: chatID,
		Send:   make(chan []byte, 10),
	}

	hub.Register() <- client
	time.Sleep(100 * time.Millisecond)

	testMsg := map[string]string{"text": "hello"}
	hub.BroadcastToRoom(chatID, testMsg)

	select {
	case msg := <-client.Send:
		var received map[string]string
		err := json.Unmarshal(msg, &received)
		assert.NoError(t, err)
		assert.Equal(t, "hello", received["text"])
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}
}

func TestWallHub_Unregister(t *testing.T) {
	hub := websocket.NewWallHub()
	go hub.Run()

	chatID := uuid.New()
	client := &websocket.WallClient{
		ID:     uuid.New(),
		ChatID: chatID,
		Send:   make(chan []byte, 10),
	}

	hub.Register() <- client
	time.Sleep(50 * time.Millisecond)

	hub.Unregister() <- client
	time.Sleep(50 * time.Millisecond)

	// After unregister, channel should be closed
	_, ok := <-client.Send
	assert.False(t, ok, "client.Send channel should be closed")

	// Trying to broadcast to this room should not send anything to anyone
	testMsg := map[string]string{"text": "hello"}
	hub.BroadcastToRoom(chatID, testMsg)
	// No panic, just works
}

func TestWallHub_BroadcastMultipleClients(t *testing.T) {
	hub := websocket.NewWallHub()
	go hub.Run()

	chatID := uuid.New()
	numClients := 5
	clients := make([]*websocket.WallClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = &websocket.WallClient{
			ID:     uuid.New(),
			ChatID: chatID,
			Send:   make(chan []byte, 10),
		}
		hub.Register() <- clients[i]
	}
	time.Sleep(100 * time.Millisecond)

	testMsg := map[string]string{"text": "broadcast"}
	hub.BroadcastToRoom(chatID, testMsg)

	for i := 0; i < numClients; i++ {
		select {
		case msg := <-clients[i].Send:
			var received map[string]string
			err := json.Unmarshal(msg, &received)
			assert.NoError(t, err)
			assert.Equal(t, "broadcast", received["text"])
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timeout waiting for message for client %d", i)
		}
	}
}

func TestWallHub_BroadcastDifferentRooms(t *testing.T) {
	hub := websocket.NewWallHub()
	go hub.Run()

	chatID1 := uuid.New()
	chatID2 := uuid.New()

	client1 := &websocket.WallClient{
		ID:     uuid.New(),
		ChatID: chatID1,
		Send:   make(chan []byte, 10),
	}
	client2 := &websocket.WallClient{
		ID:     uuid.New(),
		ChatID: chatID2,
		Send:   make(chan []byte, 10),
	}

	hub.Register() <- client1
	hub.Register() <- client2
	time.Sleep(100 * time.Millisecond)

	// Send to room 1
	hub.BroadcastToRoom(chatID1, "msg1")
	select {
	case msg := <-client1.Send:
		var received string
		json.Unmarshal(msg, &received)
		assert.Equal(t, "msg1", received)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for message for client 1")
	}

	// Verify client 2 didn't receive it
	select {
	case msg := <-client2.Send:
		t.Fatalf("client 2 should not receive message from room 1, got: %s", string(msg))
	default:
		// OK
	}
}

func TestWallHub_SlowClient(t *testing.T) {
	hub := websocket.NewWallHub()
	go hub.Run()

	chatID := uuid.New()
	// Client with channel buffer 0 to simulate full channel
	client := &websocket.WallClient{
		ID:     uuid.New(),
		ChatID: chatID,
		Send:   make(chan []byte),
	}

	hub.Register() <- client
	time.Sleep(100 * time.Millisecond)

	// Broadcast should try to send, fail, and unregister the client
	hub.BroadcastToRoom(chatID, "msg")
	time.Sleep(100 * time.Millisecond)

	// Client should be unregistered and channel closed
	_, ok := <-client.Send
	assert.False(t, ok, "client.Send channel should be closed due to slow receiver")
}

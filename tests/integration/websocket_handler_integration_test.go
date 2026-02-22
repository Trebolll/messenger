package integration

import (
	"encoding/json"
	"messenger/internal/service/websocket"
	"messenger/internal/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	gw "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test_secret_key"

func generateValidTokenForWS(userID uuid.UUID) string {
	token, _ := utils.GenerateJWT(userID, testJWTSecret, 24*time.Hour)
	return token
}

func TestHubSendToUser(t *testing.T) {
	hub := websocket.NewHub()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     userID,
		UserID: userID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	hub.Clients[userID] = client

	msg := websocket.Message{
		Type: "test_message",
		Content: map[string]string{
			"text": "hello",
		},
	}

	hub.SendToUser(userID, msg)

	select {
	case data := <-client.Send:
		var received websocket.Message
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, "test_message", received.Type)
		assert.Equal(t, "hello", received.Content.(map[string]interface{})["text"])
	case <-time.After(1 * time.Second):
		t.Fatal("Should receive message")
	}
}

func TestHubIsUserOnline(t *testing.T) {
	hub := websocket.NewHub()

	userID := uuid.New()
	assert.False(t, hub.IsUserOnline(userID), "User should not be online initially")

	client := &websocket.Client{
		ID:     userID,
		UserID: userID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	hub.Clients[userID] = client

	assert.True(t, hub.IsUserOnline(userID), "User should be online after adding to hub")
}

func TestHubSendToNonExistentUser(t *testing.T) {
	hub := websocket.NewHub()

	nonExistentUserID := uuid.New()
	msg := websocket.Message{
		Type:    "test",
		Content: "test content",
	}

	hub.SendToUser(nonExistentUserID, msg)

	assert.False(t, hub.IsUserOnline(nonExistentUserID), "Nonexistent user should not be online")
}

func TestHubMultipleClients(t *testing.T) {
	hub := websocket.NewHub()

	clientCount := 5
	var clients []*websocket.Client

	for i := 0; i < clientCount; i++ {
		userID := uuid.New()
		client := &websocket.Client{
			ID:     userID,
			UserID: userID,
			Conn:   &gw.Conn{},
			Send:   make(chan []byte, 256),
		}
		clients = append(clients, client)
		hub.Clients[userID] = client
	}

	assert.Equal(t, clientCount, len(hub.Clients), "Should have 5 clients registered")
}

func TestHubSendToBusyClient(t *testing.T) {
	hub := websocket.NewHub()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     userID,
		UserID: userID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 1),
	}

	hub.Clients[userID] = client

	msg1 := websocket.Message{
		Type:    "msg1",
		Content: "first",
	}
	data1, _ := json.Marshal(msg1)

	select {
	case client.Send <- data1:
		t.Log("First message sent to send channel")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Failed to send first message")
	}

	msg2 := websocket.Message{
		Type:    "msg2",
		Content: "second",
	}

	hub.SendToUser(userID, msg2)

	time.Sleep(50 * time.Millisecond)

	assert.False(t, hub.IsUserOnline(userID), "Client should be removed when Send channel is full")
}

func TestHubClientState(t *testing.T) {
	hub := websocket.NewHub()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     userID,
		UserID: userID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	assert.NotNil(t, client.Send, "Send channel should be initialized")
	assert.Equal(t, userID, client.UserID, "UserID should be set")

	hub.Clients[userID] = client

	registeredClient := hub.Clients[userID]
	assert.NotNil(t, registeredClient, "Client should be in hub")
	assert.Equal(t, userID, registeredClient.UserID, "UserID should match")
}

func TestHubBroadcastToMultipleClients(t *testing.T) {
	hub := websocket.NewHub()
	go func() {
		for range hub.Broadcast {
		}
	}()

	user1ID := uuid.New()
	user2ID := uuid.New()

	client1 := &websocket.Client{
		ID:     user1ID,
		UserID: user1ID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	client2 := &websocket.Client{
		ID:     user2ID,
		UserID: user2ID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	hub.Clients[user1ID] = client1
	hub.Clients[user2ID] = client2

	msg := websocket.Message{
		Type: "broadcast",
		Content: map[string]string{
			"message": "to all",
		},
	}

	data, _ := json.Marshal(msg)

	for _, client := range hub.Clients {
		select {
		case client.Send <- data:
		default:
		}
	}

	select {
	case data1 := <-client1.Send:
		var received1 websocket.Message
		json.Unmarshal(data1, &received1)
		assert.Equal(t, "broadcast", received1.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Client1 should receive broadcast")
	}

	select {
	case data2 := <-client2.Send:
		var received2 websocket.Message
		json.Unmarshal(data2, &received2)
		assert.Equal(t, "broadcast", received2.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Client2 should receive broadcast")
	}
}

func TestHubClientRemoval(t *testing.T) {
	hub := websocket.NewHub()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     userID,
		UserID: userID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	hub.Clients[userID] = client
	assert.True(t, hub.IsUserOnline(userID), "User should be online")

	msg := websocket.Message{
		Type:    "test",
		Content: "test",
	}

	hub.SendToUser(userID, msg)

	select {
	case <-client.Send:
		t.Log("Message received, client is still online")
	case <-time.After(100 * time.Millisecond):
		t.Log("Client offline or no message")
	}
}

func TestJWTTokenGeneration(t *testing.T) {
	userID := uuid.New()
	token := generateValidTokenForWS(userID)

	assert.NotEmpty(t, token, "Token should not be empty")

	claims, err := utils.VerifyJWT(token, testJWTSecret)
	require.NoError(t, err, "Token should be valid")
	assert.Equal(t, userID, claims.UserID, "Token should contain correct UserID")
}

func TestJWTTokenExpiration(t *testing.T) {
	userID := uuid.New()
	expiredToken, _ := utils.GenerateJWT(userID, testJWTSecret, -1*time.Hour)

	_, err := utils.VerifyJWT(expiredToken, testJWTSecret)
	assert.Error(t, err, "Expired token should fail verification")
}

func TestJWTTokenWithWrongSecret(t *testing.T) {
	userID := uuid.New()
	token := generateValidTokenForWS(userID)

	_, err := utils.VerifyJWT(token, "wrong_secret")
	assert.Error(t, err, "Token with wrong secret should fail verification")
}

func TestHubMultipleMessages(t *testing.T) {
	hub := websocket.NewHub()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     userID,
		UserID: userID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	hub.Clients[userID] = client

	msgCount := 5
	for i := 0; i < msgCount; i++ {
		msg := websocket.Message{
			Type:    "msg",
			Content: i,
		}
		hub.SendToUser(userID, msg)
	}

	receivedCount := 0
	timeout := time.After(2 * time.Second)

	for receivedCount < msgCount {
		select {
		case <-client.Send:
			receivedCount++
		case <-timeout:
			break
		}
	}

	assert.Equal(t, msgCount, receivedCount, "Should receive all messages")
}

func TestHubConcurrentSendOperations(t *testing.T) {
	hub := websocket.NewHub()

	user1ID := uuid.New()
	user2ID := uuid.New()

	client1 := &websocket.Client{
		ID:     user1ID,
		UserID: user1ID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	client2 := &websocket.Client{
		ID:     user2ID,
		UserID: user2ID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	hub.Clients[user1ID] = client1
	hub.Clients[user2ID] = client2

	go func() {
		for i := 0; i < 5; i++ {
			msg := websocket.Message{
				Type:    "concurrent",
				Content: i,
			}
			hub.SendToUser(user1ID, msg)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		for i := 0; i < 5; i++ {
			msg := websocket.Message{
				Type:    "concurrent",
				Content: i,
			}
			hub.SendToUser(user2ID, msg)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	receivedCount1 := 0
	receivedCount2 := 0
	timeout := time.After(2 * time.Second)

	for receivedCount1 < 5 || receivedCount2 < 5 {
		select {
		case <-client1.Send:
			receivedCount1++
		case <-client2.Send:
			receivedCount2++
		case <-timeout:
			break
		}
	}

	assert.Equal(t, 5, receivedCount1, "User1 should receive 5 messages")
	assert.Equal(t, 5, receivedCount2, "User2 should receive 5 messages")
}

func TestHubMessageSerialization(t *testing.T) {
	hub := websocket.NewHub()

	userID := uuid.New()
	client := &websocket.Client{
		ID:     userID,
		UserID: userID,
		Conn:   &gw.Conn{},
		Send:   make(chan []byte, 256),
	}

	hub.Clients[userID] = client

	complexMsg := websocket.Message{
		Type: "complex",
		Content: map[string]interface{}{
			"user_id": userID.String(),
			"text":    "hello world",
			"count":   42,
		},
	}

	hub.SendToUser(userID, complexMsg)

	select {
	case data := <-client.Send:
		var received websocket.Message
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, "complex", received.Type)
		assert.NotNil(t, received.Content)
	case <-time.After(1 * time.Second):
		t.Fatal("Should receive message")
	}
}

package service

import (
	"encoding/json"
	"messenger/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAISuggest_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test_api_key", r.Header.Get("x-api-key"))

		resp := map[string]interface{}{
			"content": []map[string]string{
				{"text": "Improved message content"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	aiService := service.NewAIServiceWithClient("test_api_key", server.URL, "claude-model", server.Client())

	req := service.AISuggestRequest{
		Text:   "Hello world",
		Action: service.AIActionImprove,
	}

	result, err := aiService.Suggest(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Improved message content", result.Result)
	assert.False(t, result.IsAdvice)
}

func TestAISuggest_APIError(t *testing.T) {
	// Mock server returning error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"error": map[string]string{
				"message": "Overloaded",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	aiService := service.NewAIServiceWithClient("test_api_key", server.URL, "claude-model", server.Client())

	req := service.AISuggestRequest{
		Text:   "Hello world",
		Action: service.AIActionImprove,
	}

	result, err := aiService.Suggest(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "anthropic error: Overloaded")
}

func TestAISuggest_ToneAction(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"content": []map[string]string{
				{"text": "Tone analysis result"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	aiService := service.NewAIServiceWithClient("test_api_key", server.URL, "claude-model", server.Client())

	req := service.AISuggestRequest{
		Text:   "Hello world",
		Action: service.AIActionTone,
	}

	result, err := aiService.Suggest(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsAdvice)
}

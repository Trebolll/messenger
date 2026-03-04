package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type AIService struct {
	apiKey     string
	apiURL     string
	model      string
	httpClient *http.Client
}

func NewAIService() *AIService {
	return &AIService{
		apiKey:     os.Getenv("ANTHROPIC_API_KEY"),
		apiURL:     "https://api.anthropic.com/v1/messages",
		model:      "claude-haiku-4-5-20251001",
		httpClient: &http.Client{},
	}
}

// NewAIServiceWithClient for testing
func NewAIServiceWithClient(apiKey, apiURL, model string, client *http.Client) *AIService {
	return &AIService{
		apiKey:     apiKey,
		apiURL:     apiURL,
		model:      model,
		httpClient: client,
	}
}

type AIAction string

const (
	AIActionImprove  AIAction = "improve"
	AIActionShorten  AIAction = "shorten"
	AIActionTone     AIAction = "tone"
	AIActionContinue AIAction = "continue"
	AIActionReply    AIAction = "reply"
)

type AISuggestRequest struct {
	Text    string   `json:"text" binding:"required"`
	Action  AIAction `json:"action" binding:"required"`
	Context string   `json:"context,omitempty"`
}

type AISuggestResponse struct {
	Result   string `json:"result"`
	IsAdvice bool   `json:"is_advice"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *AIService) Suggest(req AISuggestRequest) (*AISuggestResponse, error) {
	log.Printf("[AI] action=%s, apiKey set=%v", req.Action, s.apiKey != "")

	prompt, isAdvice := s.buildPrompt(req.Action, req.Text, req.Context)

	body, err := json.Marshal(anthropicRequest{
		Model:     s.model,
		MaxTokens: 500,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", s.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[AI] Anthropic status=%d body=%s", resp.StatusCode, string(respBody))

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if anthropicResp.Error != nil {
		return nil, fmt.Errorf("Anthropic API error: %s", anthropicResp.Error.Message)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Anthropic API")
	}

	return &AISuggestResponse{
		Result:   anthropicResp.Content[0].Text,
		IsAdvice: isAdvice,
	}, nil
}

func (s *AIService) buildPrompt(action AIAction, text string, context string) (string, bool) {
	switch action {
	case AIActionImprove:
		return fmt.Sprintf(
			"Улучши это сообщение: сделай его более вежливым, понятным и грамотным. "+
				"Верни ТОЛЬКО улучшенный текст без каких-либо пояснений и кавычек:\n%s", text,
		), false

	case AIActionShorten:
		return fmt.Sprintf(
			"Сократи это сообщение до 1-2 предложений, сохранив главный смысл. "+
				"Верни ТОЛЬКО сокращённый текст без пояснений и кавычек:\n%s", text,
		), false

	case AIActionTone:
		return fmt.Sprintf(
			"Проанализируй тон этого сообщения. Напиши 2-3 предложения: "+
				"какой тон у сообщения и как его можно улучшить. Будь конкретен:\n%s", text,
		), true

	case AIActionContinue:
		return fmt.Sprintf(
			"Предложи естественное продолжение для этого незаконченного сообщения. "+
				"Верни ТОЛЬКО продолжение (не повторяй начало) без пояснений и кавычек:\n%s", text,
		), false

	case AIActionReply:
		return fmt.Sprintf(
			"Ты — умный помощник в мессенджере. Проанализируй этот диалог и дай совет пользователю.\n\n"+
				"Диалог (последние сообщения):\n%s\n\n"+
				"Ответь на русском языке. Дай конкретный совет: что лучше ответить, "+
				"как сформулировать, или объясни почему лучше промолчать. "+
				"Если есть хороший вариант ответа — предложи готовый текст в кавычках. "+
				"Будь кратким (2-4 предложения).", context,
		), true

	default:
		return fmt.Sprintf("Улучши это сообщение:\n%s", text), false
	}
}

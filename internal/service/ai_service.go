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

// ═══ Агенты ИИ ═══════════════════════════════════════════════════════════════

type AIAgent string

const (
	AgentClaude AIAgent = "claude" // Anthropic Claude Haiku — платный
	AgentGroq   AIAgent = "groq"   // Groq llama-3.3-70b — бесплатный tier
	AgentGemini AIAgent = "gemini" // Google Gemini Flash — бесплатный tier
	AgentOllama AIAgent = "ollama" // Ollama локальный — полностью бесплатный
)

type AgentInfo struct {
	ID          AIAgent `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Free        bool    `json:"free"`
	Icon        string  `json:"icon"`
}

var Agents = []AgentInfo{
	{AgentClaude, "Claude Haiku", "Anthropic · быстрый и точный", false, "✦"},
	{AgentGroq, "Llama 3.3 70B", "Groq · бесплатно · очень быстро", true, "⚡"},
	{AgentGemini, "Gemini Flash", "Google · бесплатно · 1500/день", true, "◆"},
	{AgentOllama, "Ollama (local)", "Локально · полностью бесплатно", true, "🖥"},
}

// ═══ Сервис ═══════════════════════════════════════════════════════════════════

type AIService struct {
	anthropicKey string
	groqKey      string
	geminiKey    string
	ollamaURL    string
	httpClient   *http.Client
}

func NewAIService() *AIService {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	return &AIService{
		anthropicKey: os.Getenv("ANTHROPIC_API_KEY"),
		groqKey:      os.Getenv("GROQ_API_KEY"),
		geminiKey:    os.Getenv("GEMINI_API_KEY"),
		ollamaURL:    ollamaURL,
		httpClient:   &http.Client{},
	}
}

// NewAIServiceWithClient for testing
func NewAIServiceWithClient(apiKey, apiURL, model string, client *http.Client) *AIService {
	return &AIService{
		anthropicKey: apiKey,
		httpClient:   client,
	}
}

// ═══ Действия ════════════════════════════════════════════════════════════════

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
	Agent   AIAgent  `json:"agent,omitempty"`
}

type AISuggestResponse struct {
	Result   string `json:"result"`
	IsAdvice bool   `json:"is_advice"`
}

// ═══ Роутинг к провайдеру ════════════════════════════════════════════════════

func (s *AIService) Suggest(req AISuggestRequest) (*AISuggestResponse, error) {
	agent := req.Agent
	if agent == "" {
		agent = AgentClaude
	}
	log.Printf("[AI] agent=%s action=%s", agent, req.Action)

	prompt, isAdvice := buildPrompt(req.Action, req.Text, req.Context)

	var result string
	var err error

	switch agent {
	case AgentGroq:
		result, err = s.callGroq(prompt)
	case AgentGemini:
		result, err = s.callGemini(prompt)
	case AgentOllama:
		result, err = s.callOllama(prompt)
	default:
		result, err = s.callAnthropic(prompt)
	}

	if err != nil {
		return nil, err
	}
	return &AISuggestResponse{Result: result, IsAdvice: isAdvice}, nil
}

// ═══ Anthropic ═══════════════════════════════════════════════════════════════

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

func (s *AIService) callAnthropic(prompt string) (string, error) {
	body, _ := json.Marshal(anthropicRequest{
		Model:     "claude-haiku-4-5-20251001",
		MaxTokens: 500,
		Messages:  []anthropicMessage{{Role: "user", Content: prompt}},
	})
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.anthropicKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var r anthropicResponse
	json.Unmarshal(data, &r)
	if r.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", r.Error.Message)
	}
	if len(r.Content) == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}
	return r.Content[0].Text, nil
}

// ═══ Groq ════════════════════════════════════════════════════════════════════

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *AIService) callGroq(prompt string) (string, error) {
	body, _ := json.Marshal(openAIRequest{
		Model:    "llama-3.3-70b-versatile",
		Messages: []openAIMessage{{Role: "user", Content: prompt}},
	})
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.groqKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var r openAIResponse
	json.Unmarshal(data, &r)
	if r.Error != nil {
		return "", fmt.Errorf("groq error: %s", r.Error.Message)
	}
	if len(r.Choices) == 0 {
		return "", fmt.Errorf("groq: empty response")
	}
	return r.Choices[0].Message.Content, nil
}

// ═══ Google Gemini ════════════════════════════════════════════════════════════

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *AIService) callGemini(prompt string) (string, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=%s",
		s.geminiKey,
	)
	body, _ := json.Marshal(geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
	})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var r geminiResponse
	json.Unmarshal(data, &r)
	if r.Error != nil {
		return "", fmt.Errorf("gemini error: %s", r.Error.Message)
	}
	if len(r.Candidates) == 0 || len(r.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}
	return r.Candidates[0].Content.Parts[0].Text, nil
}

// ═══ Ollama (local) ═══════════════════════════════════════════════════════════

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}
type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (s *AIService) callOllama(prompt string) (string, error) {
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3"
	}
	body, _ := json.Marshal(ollamaRequest{
		Model:  ollamaModel,
		Prompt: prompt,
		Stream: false,
	})
	req, _ := http.NewRequest("POST", s.ollamaURL+"/api/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama недоступен (%s). Убедитесь что Ollama запущен локально", s.ollamaURL)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var r ollamaResponse
	json.Unmarshal(data, &r)
	if r.Error != "" {
		return "", fmt.Errorf("ollama error: %s", r.Error)
	}
	return r.Response, nil
}

// ═══ Промпты ═════════════════════════════════════════════════════════════════

func buildPrompt(action AIAction, text string, context string) (string, bool) {
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

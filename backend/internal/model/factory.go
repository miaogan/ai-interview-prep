// Package model builds the LLM chat model from environment config.
package model

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

const (
	defaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultModel   = "qwen-plus"
)

// NewChatModel creates a ToolCallingChatModel from env:
//
//	LLM_BASE_URL  OpenAI-compatible endpoint (default DashScope compatible-mode)
//	LLM_API_KEY   API key; falls back to DASHSCOPE_API_KEY
//	LLM_MODEL     model name (default qwen-plus)
//
// Missing API key is a hard error — never silently degrade.
func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("DASHSCOPE_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("缺少 LLM_API_KEY（或 DASHSCOPE_API_KEY）环境变量，无法调用真实模型")
	}
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	name := os.Getenv("LLM_MODEL")
	if name == "" {
		name = defaultModel
	}
	temp := float32(0.2)
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       name,
		Timeout:     90 * time.Second,
		Temperature: &temp,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
}

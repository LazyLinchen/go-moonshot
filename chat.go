package go_moonshot

import (
	"context"
	"net/http"
)

const (
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleSystem    = "system"
)

type ChatCompletionRequest struct {
	Messages         []Message             `json:"messages"`
	Model            string                `json:"model"`
	LogProbs         bool                  `json:"logprobs,omitempty"`
	MaxTokens        int                   `json:"max_tokens,omitempty"`
	N                int                   `json:"n,omitempty"`
	ResponseFormat   ResponseFormat        `json:"response_format,omitempty"`
	Seed             int                   `json:"seed,omitempty"`
	Temperature      NullableType[float64] `json:"temperature,omitempty"`
	TopP             NullableType[float64] `json:"top_p,omitempty"`
	PresencePenalty  float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64               `json:"frequency_penalty,omitempty"`
	Tools            []*Tool               `json:"tools,omitempty"`
	ToolChoice       ToolChoice            `json:"tool_choice,omitempty"`
	Stream           bool                  `json:"stream"`
}

type ChatCompletionResponse struct {
	Completion
}

func (c *Client) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (response ChatCompletionResponse, err error) {
	urlSuffix := "/chat/completions"
	req, err := c.newRequest(ctx, http.MethodPost, c.fullUrl(urlSuffix), withBody(request))
	if err != nil {
		return
	}
	err = c.sendRequest(req, &response)
	return
}

type ChatCompletionStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	FinishReason NullableType[string] `json:"finishReason"`
	Usage        Usage                `json:"usage"`
}

type ChatCompletionStream struct {
	*streamReader[ChatCompletionStreamResponse]
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, request ChatCompletionRequest) (stream *ChatCompletionStream, err error) {
	urlSuffix := "/chat/completions"

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Cache-Control", "no-cache")

	request.Stream = true
	req, err := c.newRequest(ctx, http.MethodPost, c.fullUrl(urlSuffix), withBody(request), withHeader(headers))
	if err != nil {
		return nil, err
	}
	resp, err := sendRequestStream[ChatCompletionStreamResponse](c, req)
	if err != nil {
		return
	}
	stream = &ChatCompletionStream{
		streamReader: resp,
	}
	return
}

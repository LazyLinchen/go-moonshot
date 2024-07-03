package go_moonshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Chunk struct {
	ID                string `json:"id"`
	Model             string `json:"model"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index        NullableType[int]    `json:"index"`
		Delta        *Message             `json:"delta"`
		FinishReason NullableType[string] `json:"finish_reason"`
		LogProbs     json.RawMessage      `json:"logprobs"`
	} `json:"choices"`
}

func (c *Chunk) GetDeltaByIndex(idx int) (delta *Message) {
	if idx > len(c.Choices)-1 {
		return new(Message)
	}
	return c.Choices[idx].Delta
}

func (c *Chunk) GetDeltaContentByIndex(idx int) string {
	delta := c.GetDeltaByIndex(idx)
	if delta.Content == nil {
		return ""
	}
	return strings.TrimSpace(delta.Content.Text)
}

func (c *Chunk) GetDeltaRoleByIndex(idx int) string {
	delta := c.GetDeltaByIndex(idx)
	return strings.TrimSpace(delta.Role)
}

func (c *Chunk) GetFinishReasonByIndex(idx int) string {
	if idx > len(c.Choices)-1 {
		return ""
	}
	if finishReason := c.Choices[idx].FinishReason; finishReason.IsNull() {
		return ""
	} else {
		return strings.TrimSpace(finishReason.Value())
	}
}

func (c *Chunk) GetToolCallsByIndex(idx int) []*ToolCall {
	delta := c.GetDeltaByIndex(idx)
	return delta.ToolCalls
}

func (c *Chunk) GetDelta() *Message {
	return c.GetDeltaByIndex(0)
}

func (c *Chunk) GetDeltaContent() string {
	return c.GetDeltaContentByIndex(0)
}

func (c *Chunk) GetDeltaRole() string {
	return c.GetDeltaRoleByIndex(0)
}

func (c *Chunk) GetFinishReason() string {
	return c.GetFinishReasonByIndex(0)
}

func (c *Chunk) GetToolCalls() []*ToolCall {
	return c.GetToolCallsByIndex(0)
}

type Stream struct {
	C     <-chan *Chunk
	error <-chan error
}

func (s *Stream) Err() error {
	select {
	case streamErr := <-s.error:
		if streamErr != nil {
			return fmt.Errorf("stream error, %w", streamErr)
		}
	default:
	}
	return nil
}

func (s *Stream) Close() error {
	for range s.C {

	}
	return s.Err()
}

func (s *Stream) CollectMessage() (message *Message) {
	var (
		messageContentBuilder      strings.Builder
		totalCallArgumentsBuilders []strings.Builder
	)
	message = new(Message)
	for chunk := range s.C {
		if role := chunk.GetDeltaRole(); role != "" {
			message.Role = role
		}
		messageContentBuilder.WriteString(chunk.GetDeltaContent())
		if toolCalls := chunk.GetToolCalls(); toolCalls != nil {
			if message.ToolCalls == nil {
				message.ToolCalls = make([]*ToolCall, 0, len(toolCalls))
			}
			for _, toolCall := range toolCalls {
				if !toolCall.Index.IsNull() {
					toolCallIndex := toolCall.Index.Value()
					if len(message.ToolCalls) <= toolCallIndex {
						for i := 0; i < toolCallIndex-len(message.ToolCalls)+1; i++ {
							message.ToolCalls = append(message.ToolCalls, new(ToolCall))
							totalCallArgumentsBuilders = append(totalCallArgumentsBuilders, strings.Builder{})
						}
					}
					messageToolCall := message.ToolCalls[toolCallIndex]
					if messageToolCall.ID == "" {
						message.ToolCalls[toolCallIndex].ID = toolCall.ID
					}
					if messageToolCall.Type == "" {
						message.ToolCalls[toolCallIndex].Type = toolCall.Type
					}
				}
			}
		}
	}
	message.Content = &Content{Text: messageContentBuilder.String()}
	for i := 0; i < len(message.ToolCalls); i++ {

	}
	return message
}

type ResponseFormat string

func (r ResponseFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": string(r),
	})
}

type Tool struct {
	Type     string         `json:"type"`
	Function json.Marshaler `json:"function"`
}

type ToolChoice string

func (t ToolChoice) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": "function",
		"function": map[string]any{
			"name": string(t),
		},
	})
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
}

func (e *Error) Error() string {
	return "moonshot: " + e.Message
}

type ImageUrl struct {
	Url    string `json:"url"`
	Detail string `json:"detail"`
}

type Part struct {
	Type     string    `json:"type"`
	Text     string    `json:"text"`
	ImageUrl *ImageUrl `json:"image_url"`
}

type Content struct {
	Text  string `json:"text"`
	Parts []*Part
}

func (c *Content) MarshalJSON() ([]byte, error) {
	if c == nil || (c.Text == "" && c.Parts == nil) {
		return json.Marshal(nil)
	}
	if c.Text != "" {
		return json.Marshal(c.Text)
	}
	return json.Marshal(c.Parts)
}

func (c *Content) UnmarshalJSON(data []byte) error {
	decoder := newDecoder(data)
	tok, _ := decoder.Token()
	if tok == nil {
		return nil
	}
	switch tokVal := tok.(type) {
	case string:
		return json.Unmarshal(data, &c.Text)
	case json.Delim:
		if tokVal == '[' {
			return json.Unmarshal(data, &c.Parts)
		}
	}
	return fmt.Errorf("cannot unmarshal %s into Go Value of type Content", tokenType(tok))
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Completion struct {
	ID                string `json:"id"`
	Model             string `json:"model"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index        int                  `json:"index"`
		Message      *Message             `json:"message"`
		FinishReason NullableType[string] `json:"finish_reason"`
		LogProbs     json.RawMessage      `json:"logprobs"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

func (c *Completion) GetMessageByIndex(idx int) (message *Message) {
	if idx > len(c.Choices)-1 {
		return new(Message)
	}
	return c.Choices[idx].Message
}

func (c *Completion) GetMessageContentByIndex(idx int) string {
	message := c.GetMessageByIndex(idx)
	if message.Content == nil {
		return ""
	}
	return strings.TrimSpace(message.Content.Text)
}

type Message struct {
	Role       string      `json:"role"`
	Content    *Content    `json:"content,omitempty"`
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []*ToolCall `json:"tool_calls,omitempty"`
}

type NullableType[T interface {
	string | int | float64 | bool
}] string

func (nullableVal NullableType[T]) IsNull() bool {
	var ty T
	switch any(ty).(type) {
	case string:
		return false
	default:
		return nullableVal == ""
	}
}

func (nullableVal NullableType[T]) Value() (val T) {
	if !nullableVal.IsNull() {
		if valPtr, isString := any(&val).(*string); isString {
			*valPtr = string(nullableVal)
			return val
		}
		if unmarshalErr := json.Unmarshal([]byte(nullableVal), &val); unmarshalErr != nil {
			return val
		}
	}
	return val
}

func (nullableVal NullableType[T]) MarshalJSON() ([]byte, error) {
	var ty T
	switch any(ty).(type) {
	case string:
		return json.Marshal(string(nullableVal))
	case int:
		if nullableVal == "" {
			return json.Marshal(nil)
		}
		intVal, parseIntErr := strconv.Atoi(string(nullableVal))
		if parseIntErr != nil {
			return nil, fmt.Errorf("cannot marshal %s into Go Value of type int", nullableVal)
		}
		return json.Marshal(intVal)
	case float64:
		if nullableVal == "" {
			return json.Marshal(nil)
		}
		floCal, parseFloErr := strconv.ParseFloat(string(nullableVal), 64)
		if parseFloErr != nil {
			return nil, fmt.Errorf("cannot marshal %s into Go Value of type float64", nullableVal)
		}
		return json.Marshal(floCal)
	case bool:
		if nullableVal == "" {
			return json.Marshal(nil)
		}
		boolVal, parseBoolErr := strconv.ParseBool(string(nullableVal))
		if parseBoolErr != nil {
			return nil, fmt.Errorf("cannot marshal %s into Go Value of type bool", nullableVal)
		}
		return json.Marshal(boolVal)
	}
	return []byte(nullableVal), nil
}

func (nullableVal *NullableType[T]) UnMarshalJSON(data []byte) error {
	decoder := newDecoder(data)
	tok, _ := decoder.Token()
	if tok == nil {
		return nil
	}
	var (
		ty         T
		assignable bool
	)
	switch tokVal := tok.(type) {
	case string:
		_, assignable = any(ty).(string)
		if assignable {
			var stringVal string
			if unmarshalErr := json.Unmarshal(data, &stringVal); unmarshalErr != nil {
				return unmarshalErr
			}
			*nullableVal = NullableType[T](stringVal)
			return nil
		}
	case json.Number:
		switch any(ty).(type) {
		case int:
			_, parseIntErr := tokVal.Int64()
			assignable = parseIntErr == nil
		case float64:
			_, parseFloatErr := tokVal.Float64()
			assignable = parseFloatErr == nil
		default:
			assignable = false
		}
	case bool:
		_, assignable = any(ty).(bool)
	}
	if assignable {
		*nullableVal = NullableType[T](data)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into Go Value of type NullableType[%T]", tokenType(tok), ty)
}

type ToolCall struct {
	Index NullableType[int] `json:"index,omitempty"`
	ID    string            `json:"id"`
	Type  string            `json:"type"`
}

func newDecoder(data []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder
}

func tokenType(tok json.Token) string {
	if tok == nil {
		return "null"
	}
	switch tokVal := tok.(type) {
	case bool:
		return "bool"
	case json.Number:
		return "number"
	case string:
		return "string"
	case json.Delim:
		switch tokVal {
		case '[':
			return "array"
		case '{':
			return "object"
		}
	}
	return "unknown"
}

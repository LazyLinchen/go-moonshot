package go_moonshot

import (
	"context"
	"testing"
)

var (
	apiKey = "sk-xxx"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(apiKey)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v\n", client)

	resp, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model: "moonshot-v1-128k",
		Messages: []Message{
			{
				Role: ChatMessageRoleUser,
				Content: &Content{
					Text:  "你好",
					Parts: nil,
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v\n", resp)
	t.Logf("%+v\n", resp.Choices[0].Message.Content.Text)
}

func TestClient_CreateChatCompletionStream(t *testing.T) {
	client, err := NewClient(apiKey)
	if err != nil {
		t.Error(err)
		return
	}
	stream, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{
		Model: "moonshot-v1-128k",
		Messages: []Message{
			{
				Role: ChatMessageRoleUser,
				Content: &Content{
					Text: "中国近代史",
				},
			},
		},
		Stream: true,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer stream.Close()
	for {
		resp, err := stream.Recv()
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("%s", resp.Choices[0].Delta.Content)
	}
}

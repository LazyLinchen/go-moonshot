# go-moonshot

This library provides unofficial Go clients for [Moonshot API](https://platform.moonshot.cn/docs/api/chat#%E5%9F%BA%E6%9C%AC%E4%BF%A1%E6%81%AF).

## Installation

```
go get github.com/LazyLinchen/go-moonshot
```
go-moonshot requires Go version 1.20 or greater.

## Usage

### Moonshot example usage:

```go
package main

import (
    "context"
    "fmt"
    gomoonshot "github.com/LazyLinchen/go-moonshot"
)

func main() {
    client, err := gomoonshot.NewClient("apiKey")
    if err != nil {
        fmt.Printf("NewClient error: %v\n", err)
        return
    }
    resp, err := client.CreateChatCompletion(context.Background(), gomoonshot.ChatCompletionRequest{
        Model: "moonshot-v1-128k",
        Messages: []gomoonshot.Message{
			{
				Role: gomoonshot.ChatMessageRoleUser,
				Content: &gomoonshot.Content{
					Text:  "你好",
					Parts: nil,
				},
			},
		},
    })

    if err != nil {
        fmt.Printf("ChatCompletion error: %v\n", err)
        return
    }

    fmt.Println(resp.Result)
}

```

### Other example:
<details>
<summary>Streaming completion</summary>

```go

package main

import (
	"context"
	"fmt"
	gomoonshot "github.com/LazyLinchen/go-moonshot"
	"net/http"
)

func main() {
	client, err := gomoonshot.NewClient("apiKey")
	if err != nil {
		fmt.Printf("NewClient error: %v\n", err)
		return
	}
	stream, err := client.CreateChatCompletionStream(context.Background(), gomoonshot.ChatCompletionRequest{
        Model: "moonshot-v1-128k",
        Messages: []gomoonshot.Message{
            {
                Role:    gomoonshot.ChatMessageRoleUser,
                Content: &gomoonshot.Content{
                    Text:  "你好",
                },
            },
        },
    })
	if err != nil {
        fmt.Printf("CreateChatCompletionStream error: %v\n", err)
        return
    }
	defer stream.Close()
	for {
		resp, err := stream.Recv()
		if err != nil {
			fmt.Printf("Recv error: %v\n", err)
            return
		}
		fmt.Println(resp.Choies[0].Delta.Content.Text)
    }
	
}

```
</details>
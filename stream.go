package go_moonshot

import (
	"bufio"
	"bytes"
	gomoonshot "github.com/LazyLinchen/go-moonshot/internal"
	"io"
	"net/http"
)

type streamable interface {
	ChatCompletionStreamResponse
}

type streamReader[T streamable] struct {
	isFinished  bool
	scanner     *bufio.Scanner
	response    *http.Response
	unmarshaler gomoonshot.Unmarshaler
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}
	response, err = stream.processLines()
	return
}

func (stream *streamReader[T]) processLines() (T, error) {
	if stream.scanner == nil {
		stream.scanner = bufio.NewScanner(stream.response.Body)
	}
	for {
		if !stream.scanner.Scan() {
			stream.isFinished = true
			return *new(T), stream.scanner.Err()
		}
		line := stream.scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var value []byte
		if i := bytes.IndexRune(line, ':'); i != -1 {
			value = line[i+1:]
			if len(value) != 0 && value[0] == ' ' {
				value = value[1:]
			}
		}

		if string(value) == "[DONE]" {
			stream.isFinished = true
			return *new(T), io.EOF
		}

		var response T
		unmarshalErr := stream.unmarshaler.Unmarshal(value, &response)
		if unmarshalErr != nil {
			return *new(T), unmarshalErr
		}
		return response, nil
	}
}

func (stream *streamReader[T]) Close() {
	_ = stream.response.Body.Close()
}

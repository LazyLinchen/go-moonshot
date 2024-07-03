package go_moonshot

import "net/http"

const (
	baseUrl   = "https://api.moonshot.cn"
	apiPrefix = "/v1"
)

type ClientConfig struct {
	ApiKey  string
	BaseUrl string

	HTTPClient *http.Client
}

func DefaultConfig(apiKey string) ClientConfig {
	return ClientConfig{
		ApiKey:     apiKey,
		BaseUrl:    baseUrl,
		HTTPClient: &http.Client{},
	}
}

func (ClientConfig) String() string {
	return "<Moonshot API ClientConfig>"
}

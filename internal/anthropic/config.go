package anthropic

type Config struct {
	baseURL string
	apiKey  string
}

func NewConfig(baseURL, apiKey string) *Config {
	return &Config{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

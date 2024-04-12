package anthropic

// Server-Sent-Events (SSE)
type SSEMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type SSEData struct {
	Type string `json:"type"`
}

type ContentBlockDelta struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"Delta"`
}

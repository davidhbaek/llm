package anthropic

// Price is shown as $USD per 1M tokens
const (
	HAIKU_INPUT_COST  = 0.25 / 1000000
	HAIKU_OUTPUT_COST = 1.25 / 1000000

	SONNET_INPUT_COST  = 3.00 / 1000000
	SONNET_OUTPUT_COST = 15.00 / 1000000

	OPUS_INPUT_COST  = 15.00 / 1000000
	OPUS_OUTPUT_COST = 75.00 / 1000000
)

// func getCost(model string, usage Usage) float64 {
// 	var cost float64
// 	switch model {
// 	case HAIKU:
// 		cost = float64(usage.InputTokens)*HAIKU_INPUT_COST + float64(usage.OutputTokens)*HAIKU_OUTPUT_COST
// 	case SONNET:
// 		cost = float64(usage.InputTokens)*SONNET_INPUT_COST + float64(usage.OutputTokens)*SONNET_OUTPUT_COST
// 	case OPUS:
// 		cost = float64(usage.InputTokens)*OPUS_INPUT_COST + float64(usage.OutputTokens)*OPUS_OUTPUT_COST
// 	}

// 	log.Printf("cost of query: $%f", cost)

// 	return cost
// }

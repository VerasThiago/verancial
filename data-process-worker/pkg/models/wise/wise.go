package wise

import "time"

type Wise struct {
	TransactionID         string    `json:"transacionId"`
	Status                string    `json:"status"`
	Direction             string    `json:"Direction"`
	CreatedOn             time.Time `json:"CreatedOn"`
	FinishedOn            time.Time `json:"FinishedOn"`
	SourceFeeAmount       float32   `json:"SourceFeeAmount"`
	SourceFeeCurrency     string    `json:"SourceFeeCurrency"`
	TargetFeeAmount       float32   `json:"TargetFeeAmount"`
	TargetFeeCurrency     string    `json:"TargetFeeCurrency"`
	SourceName            string    `json:"SourceName"`
	SourceAmountAfterFees float32   `json:"SourceAmountAfterFees"`
	SourceCurrency        string    `json:"SourceCurrency"`
	TargetName            string    `json:"TargetName"`
	TargetAmountAfterFees float32   `json:"TargetAmountAfterFees"`
	TargetCurrency        string    `json:"TargetCurrency"`
	ExchangeRate          float32   `json:"ExchangeRate"`
	Reference             string    `json:"Reference"`
	Batch                 string    `json:"Batch"`
}

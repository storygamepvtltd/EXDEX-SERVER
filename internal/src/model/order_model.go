package models

import "time"

type OrderRequest struct {
	Symbol    string  `json:"symbol"`     // e.g., BTCUSDT, ETHUSDT
	Side      string  `json:"side"`       // BUY or SELL
	OrderType string  `json:"order_type"` // LIMIT or MARKET
	Quantity  string  `json:"quantity"`
	WithTPSL  bool    `json:"with_tpsl"`
	UseAbs    bool    `json:"use_absolute"`
	TPPrice   float64 `json:"tp_price,omitempty"` // if absolute
	SLPrice   float64 `json:"sl_price,omitempty"` // if absolute
	TPMul     float64 `json:"tp_multiplier,omitempty"`
	SLMul     float64 `json:"sl_multiplier,omitempty"`
}

type PlacedOrder struct {
	ID        string    `bson:"_id,omitempty"` // MongoDB ID
	UserID    string    `bson:"user_id"`
	Symbol    string    `bson:"symbol"`
	Side      string    `bson:"side"`
	OrderType string    `bson:"order_type"`
	Quantity  string    `bson:"quantity"`
	WithTPSL  bool      `bson:"with_tpsl"`
	UseAbs    bool      `bson:"use_absolute"`
	TPPrice   float64   `bson:"tp_price,omitempty"`
	SLPrice   float64   `bson:"sl_price,omitempty"`
	TPMul     float64   `bson:"tp_multiplier,omitempty"`
	SLMul     float64   `bson:"sl_multiplier,omitempty"`
	Response  string    `bson:"response"` // Binance response string
	CreatedAt time.Time `bson:"created_at"`
}

// type Fill struct {
// 	Price           string `json:"price"`
// 	Qty             string `json:"qty"`
// 	Commission      string `json:"commission"`
// 	CommissionAsset string `json:"commissionAsset"`
// 	TradeID         int    `json:"tradeId"`
// }

// type OrderResponse struct {
// }

type Fill struct {
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
}

type OrderResponse struct {
	Symbol              string `json:"symbol"`
	OrderId             int64  `json:"orderId"`
	ClientOrderId       string `json:"clientOrderId"`
	TransactTime        int64  `json:"transactTime"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	Fills               []Fill `json:"fills"`
}

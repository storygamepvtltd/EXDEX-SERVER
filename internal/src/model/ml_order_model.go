package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       int64  `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	Status        string `json:"status"`
}

type OCOOrderResponse struct {
	OrderListID       int64  `json:"orderListId"`
	ContingencyType   string `json:"contingencyType"`
	ListStatusType    string `json:"listStatusType"`
	ListOrderStatus   string `json:"listOrderStatus"`
	ListClientOrderID string `json:"listClientOrderId"`
}

type OrderRecord struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserID          string             `bson:"user_id"`
	Symbol          string             `bson:"symbol"`
	Side            string             `bson:"side"`
	Type            string             `bson:"type"`
	Quantity        string             `bson:"quantity"`
	Price           string             `bson:"price"`
	TakeProfitPrice string             `bson:"take_profit_price"`
	StopLossPrice   string             `bson:"stop_loss_price"`
	OrderID         int64              `bson:"order_id"`
	ClientOrderID   string             `bson:"client_order_id"`
	Status          string             `bson:"status"`
	CreatedAt       primitive.DateTime `bson:"created_at"`
}

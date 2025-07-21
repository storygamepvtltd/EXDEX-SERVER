package models

import "go.mongodb.org/mongo-driver/bson"

type Filter struct {
	Sort       string
	SortOrder  int
	Limit      int
	Offset     int
	Conditions bson.M
}

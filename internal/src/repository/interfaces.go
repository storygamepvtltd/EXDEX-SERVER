package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	models "exdex/internal/src/model"
)

type RepositoryInterfaces interface {
	UpdateOne(collectionName string, filter, update bson.M, upsert bool) error
	FindOneWhere(collectionName string, filter bson.M, result interface{}) error
	FindByFilter(tableName string, obj interface{}, filter bson.M, opts *options.FindOptions) error
	Insert(collection string, document interface{}) error
	GetAll(ctx context.Context, collectionName string, filter bson.M, result interface{}) error
	FindByKeyValue(tableName string, obj interface{}, key string, value interface{}) error
	Create(tableName string, item interface{}) error
	GetByID(collectionName string, id string, result interface{}) error
	RemoveByFilter(collectionName string, filter bson.M) error
	GetByEmail(collectionName string, email string, result interface{}) error
	Count(collectionName string, filter bson.M) (int64, error) // New method
	GetAllByFiltter(
		collectionName string,
		items interface{},
		query bson.M, // 👈 custom filter
		filter models.Filter,
	) error
}

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	models "exdex/internal/src/model"
	database "exdex/server/databases"
)

type MongoDBRepository struct {
	Client *mongo.Client // MongoDB Client
	DBName string        // Database Name
}

var IRepo RepositoryInterfaces

func Init() {
	IRepo = &MongoDBRepository{}
}

func (r *MongoDBRepository) Create(collectionName string, item interface{}) error {
	// Set the CreatedAt and UpdatedAt fields to the current time
	now := time.Now()

	if ts, ok := item.(models.TimestampSetter); ok {
		ts.SetCreatedAt(now)
		ts.SetUpdatedAt(now)
	}

	collection := database.DB.Collection(collectionName)

	_, err := collection.InsertOne(context.Background(), item)
	if err != nil {
		return err
	}
	return nil
}

func (r *MongoDBRepository) GetAllByFiltter(
	collectionName string,
	items interface{},
	query bson.M, // 👈 custom filter
	filter models.Filter,
) error {
	findOptions := options.Find()
	findOptions.SetLimit(int64(filter.Limit))
	findOptions.SetSkip(int64(filter.Offset))

	if filter.Sort != "" {
		findOptions.SetSort(bson.M{filter.Sort: filter.SortOrder})
	}

	collection := database.DB.Collection(collectionName)
	cursor, err := collection.Find(context.Background(), query, findOptions)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	return cursor.All(context.Background(), items)
}

func (r *MongoDBRepository) FindOneWhere(collectionName string, filter bson.M, result interface{}) error {
	collection := database.DB.Collection(collectionName)
	err := collection.FindOne(context.Background(), filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("No documents found in collection %s for filter: %v", collectionName, filter)
			// return fmt.Errorf("no document found: %v", err)
			return err
		}
		log.Printf("Error finding document in collection %s: %v", collectionName, err)
		return err
	}
	log.Printf("Successfully found document in collection %s", collectionName)
	return nil
}

func (r *MongoDBRepository) FindByFilter(tableName string, obj interface{}, filter bson.M, opts *options.FindOptions) error {
	collection := database.DB.Collection(tableName)
	// Log the query
	queryJSON, _ := bson.Marshal(filter)
	log.Printf("[FindByFilter] Preparing to execute MongoDB query on table '%s': %s\n", tableName, queryJSON)

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		log.Printf("[FindByFilter] Error executing MongoDB query on table '%s': %s\n", tableName, err)
		return err
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Printf("[FindByFilter] Error closing cursor on table '%s': %s\n", tableName, err)
		}
	}()
	err = cursor.All(context.Background(), obj)
	if err != nil {
		log.Printf("[FindByFilter] Error fetching all results from cursor on table '%s': %s\n", tableName, err)
		return err
	}
	log.Printf("[FindByFilter] Successfully fetched results from table '%s'\n", tableName)

	return nil
}

func (r *MongoDBRepository) Insert(collection string, document interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get collection from database
	coll := database.DB.Collection(collection)

	// Insert the document
	_, err := coll.InsertOne(ctx, document)
	if err != nil {
		// Log the error with more details for debugging
		log.Printf("Error inserting document into %s: %v", collection, err)
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

func (r *MongoDBRepository) GetAll(ctx context.Context, collectionName string, filter bson.M, result interface{}) error {
	collection := database.DB.Collection(collectionName)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer func() {
		_ = cursor.Close(ctx) // closing cursor, ignoring close error
	}()

	if err := cursor.All(ctx, result); err != nil {
		return err
	}
	return nil
}

func (r *MongoDBRepository) FindByKeyValue(tableName string, obj interface{}, key string, value interface{}) error {
	collection := database.DB.Collection(tableName)
	filter := bson.M{key: value}

	// Log the query
	queryJSON, _ := json.Marshal(filter)
	log.Printf("Executing MongoDB query on table %s with filter: %s\n", tableName, queryJSON)

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error while executing Find operation on table %s: %v", tableName, err)
		return err
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Printf("Error closing cursor for table %s: %v", tableName, err)
		}
	}()

	if cursor.RemainingBatchLength() == 0 {
		log.Printf("No results found for table %s with filter: %s", tableName, queryJSON)
	}

	err = cursor.All(context.Background(), obj)
	if err != nil {
		log.Printf("Error decoding results for table %s: %v", tableName, err)
		return err
	}

	resultJSON, _ := json.Marshal(obj)
	log.Printf("Query result for table %s: %s", tableName, resultJSON)

	return nil
}

func (r *MongoDBRepository) GetByID(collectionName string, id string, result interface{}) error {
	collection := database.DB.Collection(collectionName)

	filter := bson.M{"id": id}
	err := collection.FindOne(context.Background(), filter).Decode(result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("no document found with ID: %s", id)
		}
		log.Printf("Error finding document in %s: %v", collectionName, err)
		return err
	}

	return nil
}

func (r *MongoDBRepository) GetByEmail(collectionName string, email string, result interface{}) error {
	collection := database.DB.Collection(collectionName)

	filter := bson.M{"email": email} // filter by email field
	err := collection.FindOne(context.Background(), filter).Decode(result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("no document found with email: %s", email)
		}
		log.Printf("Error finding document in %s: %v", collectionName, err)
		return err
	}

	return nil
}

func (r *MongoDBRepository) RemoveByFilter(collectionName string, filter bson.M) error {
	collection := database.DB.Collection(collectionName)

	queryJSON, _ := json.Marshal(filter)
	log.Printf("Executing MongoDB deletion query: %s\n", queryJSON)

	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no matching record")
	}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoDBRepository) Count(collectionName string, filter bson.M) (int64, error) {
	collection := database.DB.Collection(collectionName)
	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Printf("[Count] Error counting documents in %s: %v", collectionName, err)
		return 0, err
	}
	log.Printf("[Count] Successfully counted documents in %s: %d", collectionName, count)
	return count, nil
}

func (r *MongoDBRepository) UpdateOne(collectionName string, filter, update bson.M, upsert bool) error {
	collection := database.DB.Collection(collectionName)
	opts := options.Update().SetUpsert(upsert)

	_, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		log.Printf("[UpdateOne] Error updating document in %s: %v", collectionName, err)
		return err
	}

	log.Printf("[UpdateOne] Successfully updated document in %s", collectionName)
	return nil
}

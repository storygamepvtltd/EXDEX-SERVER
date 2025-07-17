package indexing

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateUserIndexes(ctx context.Context, userCollection *mongo.Collection) error {
	// Unique index for Email
	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_email"),
	}

	// Unique index for ExdexUserID
	exdexIDIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "exdex_user_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_exdex_user_id"),
	}

	_, err := userCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{emailIndex, exdexIDIndex})
	if err != nil {
		return err
	}

	log.Println("✅ User collection indexes created.")
	return nil
}

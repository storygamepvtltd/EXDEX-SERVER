package indexing

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

func InitIndexining(ctx context.Context, client *mongo.Client) {
	UserCollection := client.Database("exdex").Collection("users")

	// 👉 Call index creation
	if err := CreateUserIndexes(ctx, UserCollection); err != nil {
		log.Fatal("❌ Index creation failed:", err)
	}
	log.Println("✅ MongoDB connected and indexes ensured.")

}

package database

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"exdex/indexing"
)

var once sync.Once
var DB *mongo.Database

func Init() *mongo.Database {
	once.Do(func() {

		uri := viper.GetString("mongo.uri")
		dbname := viper.GetString("mongo.dbname")
		username := viper.GetString("mongo.username")
		password := viper.GetString("mongo.password")

		clientOptions := options.Client().ApplyURI(uri)
		clientOptions.SetAuth(options.Credential{
			Username: username,
			Password: password,
		})

		// Create a context with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Connect to MongoDB
		client, err := mongo.Connect(ctx, clientOptions)

		if err != nil {
			log.Panic().Msgf("Error connecting to the database at %s: %v", uri, err)
			panic(err)
		}

		// Check the connection
		err = client.Ping(ctx, nil)

		if err != nil {
			log.Panic().Msgf("Error pinging the database at %s: %v", uri, err)
			panic(err)
		}

		// Log successful connection

		log.Info().Msgf("Successfully established connection to %s/%s", uri, dbname)
		// index := indexing.Indexining{}
		indexing.InitIndexining(ctx, client)
		// index.CreatIndex(client.Database(dbname))
		DB = client.Database(dbname)

	})
	return DB
}

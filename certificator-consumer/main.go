// This is the simple golang kafka consumer app that get file content from kafka and saves it to minio:9000 public bucket (certificator)
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/go-redis/redis/v8"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"
)

func main() {
	// Init kafka reader
	kafkaReader := initKafkaReader()

	// Init minio client
	minioClient := initMinioClient()

	// Init redis client
	redisClient := initRedisClient()

	log.Println("Starting kafka consumer and starting to listen to kafka topic")

	// Consume kafka message
	consumeMessage(kafkaReader, minioClient, redisClient)
}

func initKafkaReader() *kafka.Reader {
	// Init kafka reader
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   os.Getenv("KAFKA_TOPIC"),
		GroupID: "certificator",
	})

	return kafkaReader
}

func initMinioClient() *minio.Client {
	// Init minio client
	minioClient, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
		Secure: false,
	})

	if err != nil {
		log.Fatalln(err)
	}

	return minioClient
}

// add redis to check if file already exists
func initRedisClient() *redis.Client {
	// Init redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return redisClient
}

func consumeMessage(kafkaReader *kafka.Reader, minioClient *minio.Client, redisClient *redis.Client) {
	// Consume kafka message
	for {
		msg, err := kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Message consumed from kafka and saving to minio")

		// check cache if file already exists "file:msg.Key"
		// if exists, log and skip
		// else save to minio and set cache "file:msg.Key"
		exists, err := redisClient.Exists(context.Background(), "file:"+string(msg.Key)).Result()
		if err != nil {
			log.Println("Error checking redis cache")
		}

		if exists == 1 {
			log.Println("File already exists in minio skipping...")
			continue
		}

		// Save file to minio
		err = saveFileToMinio(minioClient, msg.Key, msg.Value)
		if err != nil {
			log.Fatalln(err)
		}

		// Set cache
		err = redisClient.Set(context.Background(), "file:"+string(msg.Key), "1", 0).Err()
		if err != nil {
			log.Println("Error setting redis cache")
		}

		fmt.Println("Message consumed from kafka and saved to minio")
	}
}

func saveFileToMinio(minioClient *minio.Client, fileName []byte, fileContent []byte) error {
	// Save file to minio
	_, err := minioClient.PutObject(context.Background(),
		os.Getenv("MINIO_BUCKET"),
		string(fileName),
		bytes.NewReader(fileContent),
		int64(len(fileContent)),
		minio.PutObjectOptions{
			ContentType: "application/pdf",
		})

	if err != nil {
		log.Fatalln(err)
		return err
	}

	fmt.Println("File saved to minio")
	return nil
}

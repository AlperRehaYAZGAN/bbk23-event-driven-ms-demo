package main

// This is the simple golang gin kafka producer app server (segment-io)
// That push custom bb.event.course.completed event on 'GET /user/:user_id/course/completed' endpoint

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

func main() {
	// Init gin server
	r := gin.Default()

	// Init kafka writer
	kafkaWriter := initKafkaWriter()

	// GET /course/completed
	r.GET("/user/:user_id/course/completed", func(c *gin.Context) {
		// Get user_id from url param
		userID := c.Param("user_id")

		// Push event to kafka
		err := pushEventToKafka(c, kafkaWriter, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to push event to kafka",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Event pushed to kafka",
		})
	})

	// Run gin server on port OS.Getenv("PORT")
	r.Run(":" + os.Getenv("PORT"))
}

func initKafkaWriter() *kafka.Writer {
	// Init kafka writer
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(os.Getenv("KAFKA_BROKER")),
		Topic:    os.Getenv("KAFKA_TOPIC"),
		Balancer: &kafka.LeastBytes{},
	}

	return kafkaWriter
}

func pushEventToKafka(c *gin.Context, kafkaWriter *kafka.Writer, userID string) error {
	// Push event to kafka
	err := kafkaWriter.WriteMessages(
		c.Request.Context(),
		kafka.Message{
			Key:   []byte(userID),
			Value: []byte("bb.event.course.completed"),
		},
	)

	if err != nil {
		log.Println("Failed to push event to kafka")
		return err
	}

	fmt.Println("Event pushed to kafka")
	return nil
}

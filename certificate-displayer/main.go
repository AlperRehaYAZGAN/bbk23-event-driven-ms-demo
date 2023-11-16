// this is the simple http golang gin server that returns file from minio public bucket with "GET /certificate/:user_id"
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// Init gin server
	r := gin.Default()

	// gin debug mode
	gin.SetMode(gin.DebugMode)

	// Init minio client
	minioClient := initMinioClient()

	// GET /certificate/:user_id
	r.GET("/certificate/:user_id", func(c *gin.Context) {
		// Get user_id from url param
		userID := c.Param("user_id")

		// Get file from minio
		fileContent, err := getFileFromMinio(minioClient, userID)
		if err != nil {
			c.JSON(404, gin.H{
				"message": "File not found",
			})
			return
		}

		c.Data(200, "application/pdf", fileContent)
	})

	// Run gin server on port 8080
	r.Run(":" + os.Getenv("PORT"))
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

func getFileFromMinio(minioClient *minio.Client, fileName string) ([]byte, error) {
	// Get file from minio
	obj, err := minioClient.GetObject(context.Background(),
		os.Getenv("MINIO_BUCKET"),
		fileName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	fileContent, err := ioutil.ReadAll(obj)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	fmt.Println("File retrieved from minio")
	return fileContent, nil
}

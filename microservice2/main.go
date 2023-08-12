package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Transaction struct model
type Transaction struct {
	ID        string    `bson:"_id"`
	Timestamp time.Time `bson:"timestamp"`
	Matched   bool      `bson:"matched"`
	File1     string    `bson:"file1"`
	File2     string    `bson:"file2"`
}

const (
	SERVICE1_URL       = "SERVICE1_URL"
	AZURE_ACCOUNT_NAME = "AZURE_ACCOUNT_NAME"
	AZURE_ACCESS_KEY   = "AZURE_ACCESS_KEY"
	CONTAINER_NAME     = "CONTAINER_NAME"
	MONGO_STRING       = "MONGO_STRING"
)

var (
	service1Url   = goDotEnvVariable(SERVICE1_URL)
	accountName   = goDotEnvVariable(AZURE_ACCOUNT_NAME)
	accountKey    = goDotEnvVariable(AZURE_ACCESS_KEY)
	containerName = goDotEnvVariable(CONTAINER_NAME)
	mongoUrl      = goDotEnvVariable(MONGO_STRING)
)

func goDotEnvVariable(key string) string {

	// Load the .env file in the current directory
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func main() {

	r := gin.Default()

	r.POST("/face-recognition", faceRecognition)
	r.GET("/", rootPage)

	r.Run(":8080")
}

// Root Home page for the API
func rootPage(c *gin.Context) {
	c.JSON(200, gin.H{"Hi!": "This is a Go Dev server for face recognition API"})
}

func faceRecognition(c *gin.Context) {
	// Handling panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic occurred: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
	}()

	// Parsing the image files from the request
	_file1, err := c.FormFile("file1")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file1 is required"})
		return
	}

	_file2, err := c.FormFile("file2")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file2 is required"})
		return
	}

	// Calling microservice 1 to compare faces
	matched, err := callMicroService1(_file1, _file2)
	if err != nil {
		log.Printf("Error calling MicroService 1: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error calling MicroService 1"})
		return
	}

	// Store transaction details in MongoDB
	recordID := uuid.New()
	transaction := Transaction{
		ID:        recordID.String(),
		Timestamp: time.Now(),
		Matched:   matched,
		File1:     recordID.String() + "_" + _file1.Filename,
		File2:     recordID.String() + "_" + _file2.Filename,
	}
	var uploadFiles = []*multipart.FileHeader{_file1, _file2}
	// Uploading images to Azure Cloud Storage - Blob storage
	if err := uploadFilesToBlobStorage(uploadFiles, transaction); err != nil {
		log.Printf("Error uploading the files to blob: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save images"})
		return
	}

	// Handling errors from storing transaction
	if err := storeTransaction(transaction); err != nil {
		log.Printf("Error storing transaction: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error storing transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction recorded and saved images", "matched": matched})
}

func uploadFilesToBlobStorage(files []*multipart.FileHeader, transaction Transaction) error {
	// Creating client for storage
	client, err := storage.NewBasicClient(accountName, accountKey)
	if err != nil {
		return err
	}

	blobService := client.GetBlobService()
	// Getting container reference
	container := blobService.GetContainerReference(containerName)
	// To create the container if does not exist
	_, err = container.CreateIfNotExists(nil)
	if err != nil {
		return err
	}

	// traversing through the files to upload them in to the storage blob
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		blob := container.GetBlobReference(transaction.ID + "_" + fileHeader.Filename)

		err = blob.CreateBlockBlobFromReader(file, nil)
		if err != nil {
			return err
		}

		fmt.Printf("Uploaded %s to Azure Blob Storage\n", fileHeader.Filename)
	}

	return nil
}

func callMicroService1(file1, file2 *multipart.FileHeader) (bool, error) {
	client := http.Client{}

	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file1 and file2 to the request
	file1Part, err := writer.CreateFormFile("file1", file1.Filename)
	if err != nil {
		return false, err
	}
	file1Content, err := file1.Open()
	if err != nil {
		return false, err
	}
	defer file1Content.Close()
	_, err = io.Copy(file1Part, file1Content)
	if err != nil {
		return false, err
	}

	file2Part, err := writer.CreateFormFile("file2", file2.Filename)
	if err != nil {
		return false, err
	}
	file2Content, err := file2.Open()
	if err != nil {
		return false, err
	}
	defer file2Content.Close()
	_, err = io.Copy(file2Part, file2Content)
	if err != nil {
		return false, err
	}

	// Close the writer to finalize the form data
	writer.Close()

	// Prepare the request
	req, err := http.NewRequest("POST", service1Url, body)
	if err != nil {
		return false, err
	}

	// Set the content type header with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the request
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	// Parse the JSON response
	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return false, err
	}

	matched, ok := response["matched"].(bool)
	if !ok {
		return false, fmt.Errorf("distance not found in response")
	}

	return matched, nil
}

func storeTransaction(transaction Transaction) error {
	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUrl))
	if err != nil {
		return fmt.Errorf("error connecting to MongoDB: %w", err)
	}
	defer client.Disconnect(context.Background())

	// Access the database and collection
	db := client.Database("transactions")
	coll := db.Collection("records")

	// Insert the transaction record
	_, err = coll.InsertOne(context.Background(), transaction)
	if err != nil {
		return fmt.Errorf("error inserting transaction: %w", err)
	}

	return nil
}

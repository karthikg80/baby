package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	s3Session *s3.S3
	bucketName = os.Getenv("AWS_BUCKET_NAME") // Get S3 bucket name from environment variable
	region     = os.Getenv("AWS_REGION")      // Get AWS region from environment variable
	accessKey  = os.Getenv("AWS_ACCESS_KEY_ID") // Get AWS access key from environment variable
	secretKey  = os.Getenv("AWS_SECRET_ACCESS_KEY") // Get AWS secret key from environment variable
)

func init() {
	if region == "" {
		log.Fatalf("Missing AWS_REGION environment variable")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}
	s3Session = s3.New(sess)
}

func main() {
	r := gin.Default()

	// Serve HTML template
	r.LoadHTMLGlob("templates/*")

	// Home page route
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Upload photo route
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Could not read uploaded file: %v", err))
			return
		}

		// Open the uploaded file
		f, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Could not open uploaded file: %v", err))
			return
		}
		defer f.Close()

		// Upload to S3
		uploadInput := &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    aws.String("public-read"),
		}

		_, err = s3Session.PutObject(uploadInput)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to S3: %v", err))
			return
		}

		c.Redirect(http.StatusSeeOther, "/")
	})

	// Run the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Check if the port is already in use, use a different one if necessary
	err := r.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		log.Printf("Port %s is already in use, switching to port 8081", port)
		r.Run(":8081")
	}
}

/* Folder Structure Example:
   - main.go
   - templates/
	   - index.html
*/

/* index.html Template Example (Put this in 'templates/index.html'): 
<!DOCTYPE html>
<html>
<head>
	<title>Baby Photos</title>
</head>
<body>
	<h1>Welcome to Baby Photos!</h1>
	<form method="POST" action="/upload" enctype="multipart/form-data">
		<input type="file" name="file" accept="image/*" required>
		<button type="submit">Upload Photo</button>
	</form>
</body>
</html>
*/
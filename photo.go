package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type photo struct {
	ID        string
	UserID    string
	Filename  string
	Caption   string
	CreatedAt time.Time
	Likes     uint
}

const thumbnailSize uint = 600

var bucketName string

func init() {

	log.Info("Initializing S3")

	log.Info("Loading configuration")
	viper.SetConfigName("config") // config.toml
	viper.AddConfigPath(".")      // use working directory

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("error reading config file, %v", err)
		return
	}

	bucketName = viper.GetString("s3.bucketName")

	log.Info("S3 bucket: ", bucketName)
}

// FetchAllPhotos gets all photos for all users
func FetchAllPhotos(c *gin.Context) {
	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)

	if uid == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := findUserByID(uid.(string))

	if err != nil {
		log.Error("Could not find user:", err)
	}

	scanInput := &dynamodb.ScanInput{
		TableName: aws.String("PhotosAppPhotos"),
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	so, err := svc.Scan(scanInput)
	if err != nil {
		log.Errorf("Error querying PhotosAppPhotos: %v", err)
	}

	photos := []photo{}
	err = dynamodbattribute.UnmarshalListOfMaps(so.Items, &photos)
	if err != nil {
		log.Errorf("failed to unmarshal Query result items, %v", err)
	}

	currentUser, _ := findUserByID(uid.(string))

	c.HTML(http.StatusOK, "photos.html", gin.H{
		"user":        user,
		"photos":      photos,
		"CurrentUser": currentUser,
	})
}

// FetchSinglePhoto gets a single photo by ID
func FetchSinglePhoto(c *gin.Context) {
	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)

	if uid == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	id := c.Params.ByName("id")

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppPhotos"),
		Limit:     aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"ID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(id),
					},
				},
			},
		},
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	qo, err := svc.Query(queryInput)
	if err != nil {
		fmt.Printf("Error querying single photo: %v", err)
	}

	photos := []photo{}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &photos)
	if err != nil {
		fmt.Printf("failed to unmarshal Query result items, %v", err)
	}

	if len(photos) == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	photo := photos[0]

	log.Debug("Photo: ", photo)

	// Load user info
	user, err := findUserByID(photo.UserID)

	if err != nil {
		log.Error("Could not find user:", err)
	}

	// Load comments

	comments, err := findCommentsByPhoto(photo.ID)
	currentUser, _ := findUserByID(photo.UserID)

	c.HTML(http.StatusOK, "photo.html", gin.H{
		"user":        user,
		"photo":       photo,
		"comments":    comments,
		"CurrentUser": currentUser,
	})
}

// CreatePhoto saves the file to disk, generates its thumbnails, and stores
// metadata in the database.
func CreatePhoto(c *gin.Context) {
	sessionStore := sessions.Default(c)
	jwt := sessionStore.Get(accessToken)
	cog := NewCognito()
	sub, err := cog.ValidateToken(jwt.(string))

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Could not find user: %s", sub))
		return
	}

	form, err := c.MultipartForm()

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	file, header, err := c.Request.FormFile("photofile")

	if err != nil {
		log.Errorf("Error uploading file %v", err)
		return
	}

	defer file.Close()

	caption := form.Value["caption"][0]
	log.Info("Caption:", caption)

	// Upload file to S3 bucket
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	uploader := s3manager.NewUploader(sess)

	key := sub + "/" + header.Filename

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(mime.TypeByExtension(filepath.Ext(header.Filename))),
	})

	if err != nil {
		log.Errorf("Unable to upload file %q, %v", header.Filename, err)
		c.String(http.StatusBadRequest, fmt.Sprintf("Upload file err: %s", err.Error()))
		return
	}

	log.Info("Uploaded file:", header.Filename)

	// Insert DB record for photo and user

	photoid, err := insertPhoto(sub, header.Filename, caption)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Insert photo err: %s", err.Error()))
		return
	}

	// // Generate thumbnail

	// err = generateThumbnail(sess, sub, header.Filename, key, thumbnailSize)

	// if err != nil {
	// 	c.String(http.StatusBadRequest, fmt.Sprintf("Error generating thumbnail: %s", err.Error()))
	// 	return
	// }

	c.Redirect(http.StatusFound, fmt.Sprintf("/photos/%s", photoid))
}

// DeletePhoto deletes a single photo by ID
func DeletePhoto(c *gin.Context) {

	id := c.Params.ByName("id")

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	_, err := svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("PhotosAppPhotos"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(id)},
		},
	})

	if err != nil {
		log.Errorf("failed to delete record from DynamoDB, %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

// LikePhoto increments the 'Likes' count
func LikePhoto(c *gin.Context) {
	id := c.Params.ByName("id")

	log.Info("Liking photo: ", id)

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	result, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("PhotosAppPhotos"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(id)},
		},
		UpdateExpression: aws.String("set Likes = Likes + :num"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":num": {
				N: aws.String("1"),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	})

	if err != nil {
		log.Errorf("failed to increment PhotosAppPhotos Likes, %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	photo := photo{}
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &photo)

	if err != nil {
		log.Errorf("Unable to unmarshal response, %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"likes": photo.Likes})
}

// CommentPhoto adds a comment to a photo
func CommentPhoto(c *gin.Context) {

	id := c.Params.ByName("id")

	var comment struct {
		Comment string `json:"comment"`
	}

	if err := c.BindJSON(&comment); err != nil {
		log.Error("BindJSON error:", err.Error())
	}

	log.Printf("Comment: %v\n", comment.Comment)

	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)
	err := insertComment(id, uid.(string), comment.Comment)

	if err != nil {
		log.Error("Error inserting comment:", err.Error())
	}

	user, _ := findUserByID(uid.(string))

	c.JSON(http.StatusOK, gin.H{"username": user.Username})
}

// Insert photo record into database
func insertPhoto(uid string, fn string, caption string) (string, error) {

	id := uuid.NewV4().String()

	photo := &photo{
		ID:        id,
		UserID:    uid,
		Filename:  fn,
		Caption:   caption,
		CreatedAt: time.Now(),
	}

	av, err := dynamodbattribute.MarshalMap(photo)

	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("PhotosAppPhotos"),
		Item:      av,
	})

	if err != nil {
		log.Errorf("failed to put Record to DynamoDB, %v", err)
	}

	log.Info("Inserted photo record:", id)

	return id, nil
}

func generateThumbnail(sess *session.Session, sub string, filename string, key string, maxWidth uint) error {

	log.Infof("Fetching s3://%v/%v", bucketName, key)

	buff := &aws.WriteAtBuffer{}
	s3dl := s3manager.NewDownloader(sess)
	_, err := s3dl.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Fatalf("Could not download from S3: %v", err)
	}

	log.Infof("Decoding image")

	imageBytes := buff.Bytes()
	reader := bytes.NewReader(imageBytes)

	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Fatalf("bad response: %s", err)
	}

	log.Infof("Generating thumbnail")
	thumbnail := resize.Thumbnail(maxWidth, maxWidth, img, resize.Lanczos3)

	log.Infof("Encoding image for upload to S3")
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumbnail, nil)

	if err != nil {
		log.Errorf("JPEG encoding error: %v", err)
		return err
	}

	thumbkey := sub + "/thumb/" + filename

	log.Infof("Preparing S3 object: %s", thumbkey)

	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:        bytes.NewReader(buf.Bytes()),
		Bucket:      aws.String(bucketName),
		Key:         aws.String(thumbkey),
		ContentType: aws.String(mime.TypeByExtension(filepath.Ext(filename))),
	})

	if err != nil {
		log.Error("Failed to upload", err)
		return err
	}

	log.Println("Successfully uploaded to", result.Location)

	return nil
}

func (p *photo) TimeAgo() string {
	return humanize.Time(p.CreatedAt)
}

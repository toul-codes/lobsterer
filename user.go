package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type user struct {
	ID       string
	Email    string
	Username string
	FullName string
}

type follower struct {
	UserID     string
	FollowerID string
}

const (
	userKey     = "userid"
	accessToken = "accessToken"
)

var topicArn string

func init() {

	log.Info("Initializing SNS")

	log.Info("Loading configuration")
	viper.SetConfigName("config") // config.toml
	viper.AddConfigPath(".")      // use working directory

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("error reading config file, %v", err)
		return
	}

	topicArn = viper.GetString("sns.topicArn")

	log.Info("SNS topic: ", topicArn)
}

func loginForm(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login", "flash": flashes,
	})
}

func login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	u := &user{}
	sessionStore := sessions.Default(c)

	// Get user by username

	if _, err := findUserByUsername(username); err != nil {
		sessionStore.AddFlash("User not found")
		sessionStore.Save()
		c.HTML(http.StatusOK, "login.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  u,
		})
	} else {
		log.Info("Authenticating via Cognito: ", username)
		cog := NewCognito()
		jwt, err := cog.SignIn(username, password)

		if err != nil {
			msg := err.(awserr.Error).Message()
			log.Error("Signin Error: ", msg)
			sessionStore.AddFlash(msg)
			sessionStore.Save()
			c.HTML(http.StatusOK, "login.html", gin.H{
				"flash": sessionStore.Flashes(),
				"user":  u,
			})
		} else {
			log.Info("Authentication successful")
			sub, _ := cog.ValidateToken(jwt)
			sessionStore.Set(accessToken, jwt)
			sessionStore.Set(userKey, sub)
			sessionStore.Save()
			t := sessionStore.Get(accessToken)
			log.Debug("Testing user in session:", t)
			c.Redirect(http.StatusFound, "/photos")
		}
	}
}

func signupForm(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "signup.html", gin.H{"flash": flashes})
}

func welcome(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "welcome.html", gin.H{"flash": flashes})
}

func signup(c *gin.Context) {
	user := &user{
		FullName: c.PostForm("fullName"),
		Username: c.PostForm("username"),
		Email:    c.PostForm("email"),
	}

	sessionStore := sessions.Default(c)

	u, _ := findUserByUsername(user.Username)

	if u != nil {
		msg := "This username isn't available. Please try another."
		sessionStore.AddFlash(msg)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})
		sessionStore.Save()
		return
	}

	cog := NewCognito()
	password := c.PostForm("password")
	jwt, err := cog.SignUp(user.Username, password, user.Email, user.FullName)

	if err != nil {
		msg := err.(awserr.Error).Message()
		log.Error("SignUp error: ", msg)
		sessionStore.AddFlash(msg)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})
		sessionStore.Save()
		return
	}

	log.Info("Creating DB user:", user.Username)

	sub, err := cog.ValidateToken(jwt)

	if err != nil {
		return
	}

	log.Info("Cognito 'sub': ", sub)

	user.ID = sub // Set user ID to Cognito UUID

	// Create user in DynamoDB

	av, err := dynamodbattribute.MarshalMap(user)

	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
		sessionStore.AddFlash(err)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})

		c.Redirect(http.StatusFound, "/photos")
		sessionStore.Save()
		return
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("PhotosAppUsers"),
		Item:      av,
	})

	if err != nil {
		log.Errorf("Error: %v", err)
		sessionStore.AddFlash(err)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})
	} else {
		log.Info("Saving userid in session for: ", user.Username)
		sessionStore.Set(userKey, user.ID)
		sessionStore.Set(accessToken, jwt)
		sessionStore.Save()
		c.Redirect(http.StatusFound, "/photos")
	}

	// Send SNS notifcation

	log.Info("Sending SNS message")

	snssvc := sns.New(sess)

	var buffer bytes.Buffer
	userdata, _ := json.Marshal(user)
	buffer.Write(userdata)

	params := &sns.PublishInput{
		Subject:  aws.String("New Subscriber"),
		Message:  aws.String("User:\n" + buffer.String()),
		TopicArn: aws.String(topicArn),
	}

	_, err = snssvc.Publish(params)

	if err != nil {
		log.Errorf("Error: %v", err.Error())
	}

	sessionStore.Save()
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	c.Redirect(302, "/")
}

// Profile shows the user profile
// GET /user/:username
func Profile(c *gin.Context) {
	username := c.Params.ByName("username")

	user, err := findUserByUsername(username)

	if err != nil {
		log.Error("Error:", err)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	// Find photos by user

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppPhotos"),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(user.ID),
					},
				},
			},
		},
		IndexName: aws.String("UserID-index"),
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("Error: %v", err)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	photos := []photo{}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &photos)
	if err != nil {
		log.Errorf("failed to unmarshal Query result items, %v", err)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)
	currentUser, _ := findUserByID(uid.(string))

	c.HTML(http.StatusOK, "user.html", gin.H{
		"user":        user,
		"photos":      photos,
		"IsSelf":      uid == user.ID,
		"CurrentUser": currentUser,
	})
}

func findUserByUsername(username string) (*user, error) {

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppUsers"),
		Limit:     aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"Username": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(username),
					},
				},
			},
		},
		IndexName: aws.String("Username-index"),
	}

	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("FindUserByUsername failed: %v", err)
		return nil, err
	}

	users := []user{}
	if err := dynamodbattribute.UnmarshalListOfMaps(qo.Items, &users); err != nil {
		log.Errorf("Failed to unmarshal Query result items, %v", err)
		return nil, err
	}

	if len(users) == 0 {
		// Returned no users
		return nil, errors.New("User not found")
	}

	return &users[0], nil
}

func findUserByID(id string) (*user, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppUsers"),
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

	qo, err := svc.Query(queryInput)

	if err != nil {
		return nil, err
	}

	users := []user{}
	if err := dynamodbattribute.UnmarshalListOfMaps(qo.Items, &users); err != nil {
		log.Errorf("Failed to unmarshal Query result items, %v", err)
		return nil, err
	}

	if len(users) == 0 {
		// Returned no users
		return nil, errors.New("User not found")
	}

	return &users[0], nil
}

func (u *user) PhotoCount() uint {

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppPhotos"),
		Select:    aws.String("COUNT"),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
					},
				},
			},
		},
		IndexName: aws.String("UserID-index"),
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("Error getting photo count: %v", err)
		return 0
	}

	count := aws.Int64Value(qo.Count)

	return uint(count)
}

// Follow inserts a record into the followers table
func Follow(c *gin.Context) {
	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)
	fid := c.Params.ByName("id")

	follower := &follower{
		UserID:     fid,
		FollowerID: uid.(string),
	}

	// Insert follower into DynamoDB

	av, err := dynamodbattribute.MarshalMap(follower)

	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("PhotosAppUsers"),
		Item:      av,
	})

	if err != nil {
		log.Errorf("failed to put record to DynamoDB, %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

// Unfollow deletes a record from the followers table
func Unfollow(c *gin.Context) {
	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)
	fid := c.Params.ByName("id")

	follower := &follower{
		UserID:     fid,
		FollowerID: uid.(string),
	}

	// Delete follower from DynamoDB

	av, err := dynamodbattribute.MarshalMap(follower)

	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	_, err = svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("PhotosAppUsers"),
		Key:       av,
	})

	if err != nil {
		log.Errorf("failed to delete record from DynamoDB, %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	c.JSON(http.StatusOK, nil)
}

// Followers returns the number of followers this user has
func (u *user) Followers() uint {

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppFollowers"),
		Select:    aws.String("COUNT"),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
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
		log.Errorf("Error getting follower count: %v", err)
		return 0
	}

	count := aws.Int64Value(qo.Count)

	return uint(count)
}

// Following returns the number of users this user is following
func (u *user) Following() uint {

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppFollowers"),
		Select:    aws.String("COUNT"),
		KeyConditions: map[string]*dynamodb.Condition{
			"FollowerID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
					},
				},
			},
		},
		IndexName: aws.String("FollowerID-index"),
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("Error getting following count: %v", err)
		return 0
	}

	count := aws.Int64Value(qo.Count)

	return uint(count)
}

// Follows returns true if the user (u) follows the userid
func (u *user) Follows(userid string) bool {

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppFollowers"),
		Select:    aws.String("COUNT"),

		KeyConditions: map[string]*dynamodb.Condition{
			"FollowerID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
					},
				},
			},
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(userid),
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
		log.Errorf("Error getting follows count: %v", err)
		return false
	}

	count := aws.Int64Value(qo.Count)

	return count > 0
}

package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	log "github.com/sirupsen/logrus"
)

type comment struct {
	UserID    string
	PhotoID   string
	Text      string
	CreatedAt time.Time
}

// insertComment inserts a comment record
func insertComment(photoid string, userid string, text string) error {
	comment := &comment{
		Text:      text,
		PhotoID:   photoid,
		UserID:    userid,
		CreatedAt: time.Now(),
	}

	av, err := dynamodbattribute.MarshalMap(comment)

	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
		return err
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("PhotosAppComments"),
		Item:      av,
	})

	if err != nil {
		log.Errorf("Failed to put Record to DynamoDB, %v", err)
		return err
	}

	log.Println("Inserted comment record")

	return nil
}

// findCommentsByPhoto gets all comments for a photo
func findCommentsByPhoto(photoid string) ([]comment, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppComments"),
		KeyConditions: map[string]*dynamodb.Condition{
			"PhotoID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(photoid),
					},
				},
			},
		},
		ScanIndexForward: aws.Bool(false), // Primary sort key CreatedAt
	}

	qo, err := svc.Query(queryInput)

	if err != nil {
		return nil, err
	}

	comments := []comment{}
	if err := dynamodbattribute.UnmarshalListOfMaps(qo.Items, &comments); err != nil {
		log.Errorf("Failed to unmarshal Query result items, %v", err)
		return nil, err
	}

	return comments, nil
}

func (c *comment) Username() string {
	user, _ := findUserByID(c.UserID)
	return user.Username
}

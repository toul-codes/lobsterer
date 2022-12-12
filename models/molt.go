package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"time"
)

type Molt struct {
	ID           string `dynamodbav:"id"`
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	Created      string `dynamodbav:"created"`
	Author       string `dynamodbav:"author"`
	Content      string `dynamodbav:"content"`
	Url          string `dynamodbav:"url"`
	Deleted      bool   `dynamodbav:"deleted"`
	LikeCount    int    `dynamodbav:"like_count"`
	RemoltCount  int    `dynamodbav:"remolt_count"`
	CommentCount int    `dynamodbav:"comment_count"`
}

func (m *Molt) ById(svc ItemService, tablename string, text string) {

}

func (m *Molt) ByAuthor(svc ItemService, tablename string, text string) {

}

func (m *Molt) ByTime(svc ItemService, tablename string, text string) {

}

// Create - adds molt to db and increments user's MoltCount
func (m *Molt) Create(svc ItemService, tablename string, text string) {
	// use the iso 8601 format so that it is easier to query createdAtTime
	m.Created = fmt.Sprintf(time.Now().Format(time.RFC3339))
	// the Composite primary key is created by concatenating display to L#
	m.PK = fmt.Sprintf(PKFormat, m.ID) // M#<UserName>#<MoltId>
	m.SK = fmt.Sprintf(SKFormat, m.ID) // M#<UserName>#<MoltId>

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	_, err = svc.ItemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(tablename),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	// TODO gin flash e-mail is taken
	if err != nil {
		fmt.Printf("Couldn't add item to table: %v\n", err)
	}
	return err
}

func (m *Molt) Re(svc ItemService, tablename string, text string) {

}

func (m *Molt) Delete(svc ItemService, tablename string, text string) {

}

func (m *Molt) Edit(svc ItemService, tablename string, text string) {

}

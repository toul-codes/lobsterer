package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

type User struct {
	PK          string `dynamodbav:"PK"`
	SK          string `dynamodbav:"SK"`
	Created     string `dynamodbav:"created"`
	Name        string `dynamodbav:"name"`
	Email       string `dynamodbav:"email"`
	Display     string `dynamodbav:"display"`
	Description string `dynamodbav:"description"`
	Verified    bool   `dynamodbav:"verified"`
	Banner      string `dynamodbav:"banner"`
	Avatar      string `dynamodbav:"avatar"`
	Banned      bool   `dynamodbav:"banned"`
	Website     string `dynamodbav:"website"`
	Deleted     bool   `dynamodbav:"deleted"`
}

// AddUser adds a user to the DynamoDB table.
func AddUser(u *User, service ItemService, tablename string) error {
	item, err := attributevalue.MarshalMap(u)

	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	_, err = service.itemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tablename), Item: item,
	})
	if err != nil {
		log.Printf("Couldn't add item to table: %v\n", err)
	}
	return err
}

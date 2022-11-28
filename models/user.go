package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	PKFormat = "L#%s"
	SKFormat = "L#%s"
)

type User struct {
	PK          string `dynamodbav:"PK"`
	SK          string `dynamodbav:"SK"`
	ID          string `dynamodbav:"id"`
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

// Add - creates user record in table
func (u *User) Add(service ItemService, tablename string) error {
	// use the iso 8601 format so that it is easier to query createdAtTime
	u.Created = fmt.Sprintf(time.Now().Format(time.RFC3339))
	// the Composite primary key is created by concatenating display to L#
	u.PK = fmt.Sprintf(PKFormat, u.Display)
	u.SK = fmt.Sprintf(SKFormat, u.Display)
	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	_, err = service.itemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tablename), Item: item,
		ConditionExpression: aws.String("attribute_not_exists(email)"),
	})
	// TODO gin flash e-mail is taken
	if err != nil {
		log.Printf("Couldn't add item to table: %v\n", err)
	}
	return err
}

// Exists checks if username is already taken
func Exists(name string, service ItemService, tablename string) (bool, error) {
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf(PKFormat, name),
		"SK": fmt.Sprintf(SKFormat, name),
	}
	key, err := attributevalue.MarshalMap(selectedKeys)

	data, err := service.itemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tablename),
		Key:       key,
	},
	)
	if err != nil {
		return false, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return false, fmt.Errorf("GetItem: Data not found.\n")
	}

	return true, nil
}

// Get - retrieves a user's record from the table
func ByName(name string, svc ItemService, tablename string) (User, error) {
	user := User{}
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf(PKFormat, name),
		"SK": fmt.Sprintf(SKFormat, name),
	}
	key, err := attributevalue.MarshalMap(selectedKeys)

	data, err := svc.itemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tablename),
		Key:       key,
	},
	)
	if err != nil {
		return user, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return user, fmt.Errorf("GetItem: Data not found.\n")
	}
	err = attributevalue.UnmarshalMap(data.Item, &user)
	if err != nil {
		return user, fmt.Errorf("UnmarshalMap: %v\n", err)
	}

	return user, nil
}

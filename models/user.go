package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

const (
	PKFormat = "L#%s"
	SKFormat = "L#%s"
)

type User struct {
	ID             string `dynamodbav:"id"`
	PK             string `dynamodbav:"PK"`
	SK             string `dynamodbav:"SK"`
	GSI1PK         string `dynamodbav:"GSI1PK"`
	GSI1SK         string `dynamodbav:"GSI1SK"`
	Avatar         string `dynamodbav:"avatar"`
	Banner         string `dynamodbav:"banner"`
	Created        string `dynamodbav:"created"`
	Description    string `dynamodbav:"description"`
	Display        string `dynamodbav:"display"`
	Email          string `dynamodbav:"email"`
	Name           string `dynamodbav:"name"`
	Website        string `dynamodbav:"website"`
	Verified       bool   `dynamodbav:"verified"`
	Banned         bool   `dynamodbav:"banned"`
	Deleted        bool   `dynamodbav:"deleted"`
	Moderator      bool   `dynamodbav:"moderator"`
	FollowerCount  int    `dynamodbav:"follower_count"`
	FollowingCount int    `dynamodbav:"following_count"`
	MoltCount      int    `dynamodbav:"molt_count"`
	LikeCount      int    `dynamodbav:"like_count"`
}

// UpdateSettings - updates a string attribute associated with user
func (u *User) UpdateSettings(svc ItemService, tablename, de, w, di string) error {
	// update the user by the attrName
	if de != "" {
		err := u.UpdateStrAttr(svc, tablename, "description", de)
		if err != nil {
			fmt.Printf("ERR: %s", err)
		}
	}
	if w != "" {
		err := u.UpdateStrAttr(svc, tablename, "website", w)
		if err != nil {
			fmt.Printf("ERR: %s", err)
		}
	}
	return nil

}

// UpdateStringAttr - updates a string attribute associated with user
func (u *User) UpdateStrAttr(svc ItemService, tablename string, attr string, val string) error {
	// update the user by the attrName
	out, err := svc.ItemTable.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tablename),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: u.PK},
			"SK": &types.AttributeValueMemberS{Value: u.SK},
		},
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :%s", attr, attr)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", attr): &types.AttributeValueMemberS{Value: val},
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(out.Attributes)
	return nil
}

// Add - creates user record in table
func (u *User) Add(svc ItemService, tablename string) error {
	// use the iso 8601 format so that it is easier to query createdAtTime
	u.Created = fmt.Sprintf(time.Now().Format(time.RFC3339))
	// the Composite primary key is created by concatenating display to L#
	u.PK = fmt.Sprintf(PKFormat, u.ID) // search by id
	u.SK = fmt.Sprintf(SKFormat, u.ID)
	u.GSI1PK = u.Display                   // search by username
	u.GSI1SK = fmt.Sprintf(PKFormat, u.ID) // return the userID

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

// Delete - removes a user from Lobsterer DB & Cognito
func (u *User) Delete(svc ItemService, tablename string) error {
	return nil
}

// Exists - checks if username is already taken
func Exists(name string, svc ItemService, tablename string) (bool, error) {
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf(PKFormat, name),
		"SK": fmt.Sprintf(SKFormat, name),
	}
	key, err := attributevalue.MarshalMap(selectedKeys)

	data, err := svc.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
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

// ByID - retrieves a user's record from the table ByID
func ByID(ID string, svc ItemService, tablename string) (User, error) {
	user := User{}
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf(PKFormat, ID),
		"SK": fmt.Sprintf(SKFormat, ID),
	}
	key, err := attributevalue.MarshalMap(selectedKeys)

	data, err := svc.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tablename),
		Key:       key,
	})
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

// ByName - retrieves user's record by display name (username)
func ByName(name string, svc ItemService, tablename string) (User, error) {
	user := make([]User, 0)
	ut, err := svc.ItemTable.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(tablename),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: name},
		},
	})
	if err != nil {
		panic(err)
	}
	err = attributevalue.UnmarshalListOfMaps(ut.Items, &user)
	if err != nil {
		fmt.Errorf("UnmarshalMap: %v\n", err)
	}

	return user[0], nil
}

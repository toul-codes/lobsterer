package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	PKFormat    = "L#%s"
	SKFormat    = "L#%s"
	userKey     = "userid"
	accessToken = "accessToken"
)

type User struct {
	PK             string `dynamodbav:"PK"`
	SK             string `dynamodbav:"SK"`
	ID             string `dynamodbav:"id"`
	Created        string `dynamodbav:"created"`
	Name           string `dynamodbav:"name"`
	Email          string `dynamodbav:"email"`
	Display        string `dynamodbav:"display"`
	Description    string `dynamodbav:"description"`
	Verified       bool   `dynamodbav:"verified"`
	Banner         string `dynamodbav:"banner"`
	Avatar         string `dynamodbav:"avatar"`
	Banned         bool   `dynamodbav:"banned"`
	Website        string `dynamodbav:"website"`
	Deleted        bool   `dynamodbav:"deleted"`
	FollowerCount  int    `dynamodbav:"follower_count"`
	FollowingCount int    `dynamodbav:"following_count"`
}

type Follow struct {
	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

// Add - creates user record in table
func (u *User) Add(svc ItemService, tablename string) error {
	// use the iso 8601 format so that it is easier to query createdAtTime
	u.Created = fmt.Sprintf(time.Now().Format(time.RFC3339))
	// the Composite primary key is created by concatenating display to L#
	u.PK = fmt.Sprintf(PKFormat, u.ID)
	u.SK = fmt.Sprintf(SKFormat, u.ID)
	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	_, err = svc.itemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(tablename),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	// TODO gin flash e-mail is taken
	if err != nil {
		log.Printf("Couldn't add item to table: %v\n", err)
	}
	return err
}

// Exists - checks if username is already taken
func Exists(name string, svc ItemService, tablename string) (bool, error) {
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
		return false, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return false, fmt.Errorf("GetItem: Data not found.\n")
	}

	return true, nil
}

// ByID - retrieves a user's record from the table
func ByID(ID string, svc ItemService, tablename string) (User, error) {
	user := User{}
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf(PKFormat, ID),
		"SK": fmt.Sprintf(SKFormat, ID),
	}
	key, err := attributevalue.MarshalMap(selectedKeys)

	data, err := svc.itemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
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

// Follows - creates a user_follows_User record in db
func (u *User) Follow(svc ItemService, tablename string, fid string) error {
	// this is only for developing mode to work with GIN will need a session
	// which isn't great for running a local test programatically
	f := &Follow{
		PK: "F#" + u.ID, // my unique cognito id so can change Display as much as want
		SK: "F#" + fid,  // others unique user id so they can do the same
	}

	item, err := attributevalue.MarshalMap(f)
	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	tItems := make([]types.TransactWriteItem, 0)
	tw1 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                item,
			TableName:           aws.String(tablename),
			ConditionExpression: aws.String("attribute_not_exists(PK)"),
		},
	}
	tw2 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: u.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: u.SK,
				},
			},
			ConditionExpression: aws.String("attribute_exists(PK)"),
			TableName:           aws.String(tablename),
			UpdateExpression:    aws.String("set #following_count = #following_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#following_count": "following_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tw3 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: fmt.Sprintf(PKFormat, fid),
				},
				"SK": &types.AttributeValueMemberS{
					Value: fmt.Sprintf(SKFormat, fid),
				},
			},
			TableName:           aws.String(tablename),
			ConditionExpression: aws.String("attribute_exists(PK)"),
			UpdateExpression:    aws.String("set #follower_count = #follower_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#follower_count": "follower_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)
	tItems = append(tItems, tw3)

	_, err = svc.itemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		log.Printf("\nErr: %v", err)
	}
	return err
}

// Following - returns list of users user is following
func (u *User) Following(svc ItemService, tablename string) []User {
	following := make([]User, 0)
	p := dynamodb.NewQueryPaginator(svc.itemTable, &dynamodb.QueryInput{
		TableName:              aws.String(tablename),
		Limit:                  aws.Int32(5),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "F#" + u.ID},
		},
	})
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		err = attributevalue.UnmarshalListOfMaps(out.Items, &following)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}

	}
	for _, user := range following {
		fmt.Printf("\n%+v", user)
	}

	return following
}

// Followers - returns of users following user
func (u *User) Followers(svc ItemService, tablename string) []User {
	followers := make([]User, 0)
	p := dynamodb.NewQueryPaginator(svc.itemTable, &dynamodb.QueryInput{
		TableName:              aws.String(tablename),
		Limit:                  aws.Int32(5),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "F#" + u.ID},
		},
	})
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		err = attributevalue.UnmarshalListOfMaps(out.Items, &followers)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}

	}
	for _, user := range followers {
		fmt.Printf("\n%+v", user)
	}

	return followers
}

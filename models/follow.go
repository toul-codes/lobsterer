package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Follow struct {
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	GSI2PK string `dynamodbav:"GSI2PK"`
	GSI2SK string `dynamodbav:"GSI2SK"`
}

// Follow - creates a user_follows_User record in db
func (u *User) Follow(svc ItemService, tablename string, fid string) error {
	// this is only for developing mode to work with GIN will need a session
	// which isn't great for running a local test programatically
	f := &Follow{
		PK:     "F#" + u.ID,
		SK:     "F#" + fid,
		GSI2PK: "F#" + fid,
		GSI2SK: "F#" + u.ID,
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
	// increments how many lobsters the user is following
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

	_, err = svc.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
	}
	return err
}

// Following - who the user is following
func (u *User) Following(svc ItemService, tablename string) []User {
	following := make([]User, 0)
	p := dynamodb.NewQueryPaginator(svc.ItemTable, &dynamodb.QueryInput{
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
	return following
}

// Followers - user's followers
func (u *User) Followers(svc ItemService, tablename string) []User {
	followers := make([]User, 0)
	p := dynamodb.NewQueryPaginator(svc.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(tablename),
		Limit:                  aws.Int32(5),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :hashKey"),
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

// Unfollow - Delete following relationship & decrements
func (u *User) Unfollow(svc ItemService, tablename string, fid string) error {
	tItems := make([]types.TransactWriteItem, 0)
	// delete it from the main table
	tw1 := types.TransactWriteItem{
		Delete: &types.Delete{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: "F#" + u.ID,
				},
				"SK": &types.AttributeValueMemberS{
					Value: "F#" + fid,
				},
			},
			TableName:           aws.String(tablename),
			ConditionExpression: aws.String("attribute_exists(PK)"),
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
			UpdateExpression:    aws.String("set #following_count = #following_count - :value"),
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
			UpdateExpression:    aws.String("set #follower_count = #follower_count - :value"),
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

	_, err := svc.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
	}
	return err
}

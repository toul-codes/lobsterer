package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"testing"
)

func TestDeleteAllItems(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	err := DeleteAllItems(is.itemTable, TableName)
	if err != nil {
		log.Fatal("failed to delete all items", err)
	}

	scan, err := is.itemTable.Scan(context.TODO(), &dynamodb.ScanInput{TableName: aws.String(TableName)})
	if err != nil {
		log.Fatal("scan failed", err)
	}
	log.Printf("expected scan to have zero items; it had len=%d\n", len(scan.Items))
}

func TestNewItemService(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	if is.itemTable == nil {
		fmt.Printf("error:%v ", is)
	}
}

func TestByID(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	res, _ := ByID("3333c33b-c3cc-33bc-3333-33333e3f3f33", is, TableName)
	fmt.Printf("res: %+v", res)
}

func TestPrint(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})

	Print(is.itemTable, TableName)
}

func TestCreateTableIfNotExists(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})

	CreateTableIfNotExists(is.itemTable, TableName)

}

func TestExists(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	want := true
	got, _ := Exists("Larry", is, TableName)
	if got != want {
		fmt.Println("got", got)
	}
}

func TestUserAdd(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})

	//d := &User{
	//	ID:             "1111b11a-b1bb-11ab-1111-11111d1e1f11",
	//	Name:           "Toul",
	//	Email:          "toul@hey.com",
	//	Display:        "Toul",
	//	Description:    "I like lobsters",
	//	Verified:       false,
	//	Avatar:         "",
	//	Banner:         "",
	//	Banned:         false,
	//	Website:        "https://toul.io",
	//	Deleted:        false,
	//	FollowerCount:  0,
	//	FollowingCount: 0,
	//}
	d := &User{
		ID:             "3333c33b-c3cc-33bc-3333-33333e3f3f33",
		Name:           "Phillip",
		Email:          "phillp@gmail.com",
		Display:        "animeLover",
		Description:    "I like romcom animes",
		Verified:       false,
		Avatar:         "",
		Banner:         "",
		Banned:         false,
		Website:        "",
		Deleted:        false,
		FollowerCount:  0,
		FollowingCount: 0,
	}

	d.Add(is, TableName)

}

func TestFollow(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	phillip, _ := ByID("3333c33b-c3cc-33bc-3333-33333e3f3f33", is, TableName)
	larry := "1111b11a-b1bb-11ab-1111-11111d1e1f11"
	phillip.Follow(is, TableName, larry)
}

func TestFollowing(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	phillip, _ := ByID("3333c33b-c3cc-33bc-3333-33333e3f3f33", is, TableName)
	phillip.Following(is, TableName)
}

func TestFollowers(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	larry, _ := ByID("1111b11a-b1bb-11ab-1111-11111d1e1f11", is, TableName)
	larry.Followers(is, TableName)
}

//func TestItemService_CreateItem(t *testing.T) {
//	log.Println("starting")
//	conditionCheckFailure()
//	putItemsAndDeleteAll()
//	log.Println("completed")
//}
//
//var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
//
//func RandStringRunes(n int) string {
//	b := make([]rune, n)
//	for i := range b {
//		b[i] = letterRunes[rand.Intn(len(letterRunes))]
//	}
//	return string(b)
//}
//
//func putItemsAndDeleteAll() {
//	d := CreateLocalClient(u)
//	CreateTableIfNotExists(d, tableName)
//	for i := 0; i < 500; i++ {
//		item := map[string]types.AttributeValue{
//			"PK":     &types.AttributeValueMemberS{Value: "PK-" + strconv.Itoa(i)},
//			"SK":     &types.AttributeValueMemberS{Value: "A"},
//			"Filler": &types.AttributeValueMemberS{Value: RandStringRunes(10000)},
//		}
//		err := putItem(d, tableName, item)
//		if err != nil {
//			log.Fatal("failed to put item", err)
//		}
//	}
//
//	err := DeleteAllItems(d, tableName)
//	if err != nil {
//		log.Fatal("failed to delete all items", err)
//	}
//
//	scan, err := d.Scan(context.TODO(), &dynamodb.ScanInput{TableName: aws.String(tableName)})
//	if err != nil {
//		log.Fatal("scan failed", err)
//	}
//	log.Printf("expected scan to have zero items; it had len=%d\n", len(scan.Items))
//}
//
//func conditionCheckFailure() {
//	d := CreateLocalClient(u)
//	CreateTableIfNotExists(d, tableName)
//	err := DeleteAllItems(d, tableName)
//	if err != nil {
//		log.Fatal("failed to delete all items", err)
//	}
//	item := struct {
//		PK string `dynamodbav:"PK"`
//		SK string `dynamodbav:"SK"`
//	}{
//		PK: "ITEM#123",
//		SK: "A",
//	}
//	ddbJson, err := attributevalue.MarshalMap(item)
//	if err != nil {
//		log.Fatal("failed to marshal item", err)
//	}
//
//	log.Println("putting item")
//	err = putItem(d, tableName, ddbJson)
//	if err != nil {
//		log.Fatal("PutItem failed", err)
//	}
//
//	log.Println("putting same item, should fail with condition check failure")
//	err = putItem(d, tableName, ddbJson)
//	if err != nil {
//		log.Fatal("PutItem failed", err)
//	}
//
//	if IsConditionCheckFailure(err) {
//		log.Println("condition check failure error", err)
//	} else {
//		log.Println("general error", err)
//	}
//}
//
//func putItem(d *dynamodb.Client, tableName string, item map[string]types.AttributeValue) error {
//	_, err := d.PutItem(context.TODO(), &dynamodb.PutItemInput{
//		TableName:           aws.String(tableName),
//		Item:                item,
//		ConditionExpression: aws.String("attribute_not_exists(PK)"),
//	})
//	return err
//}

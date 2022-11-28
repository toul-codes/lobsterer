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

func TestGetByName(t *testing.T) {
	is := NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
	res, _ := ByName("Larry", is, TableName)
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

	u := &User{
		PK:          "L#Larry",
		SK:          "L#Larry",
		Name:        "Larry",
		Email:       "larry@bikinibottom.com",
		Display:     "BigLobster",
		Description: "I like weights",
		Verified:    false,
		Avatar:      "https://www.google.com/imgres?imgurl=https%3A%2F%2Fstatic.wikia.nocookie.net%2Fspongebob%2Fimages%2F0%2F05%2FLarry_the_Lobster_stock_image_standing.png%2Frevision%2Flatest%3Fcb%3D20220807062551&imgrefurl=https%3A%2F%2Fspongebob.fandom.com%2Fwiki%2FLarry_the_Lobster&tbnid=pbx86w_qC1rc5M&vet=12ahUKEwi_8Kaa48P7AhU1s2oFHWRrCPMQMygAegUIARDfAQ..i&docid=vqku3XZJ7pEJrM&w=1500&h=1500&q=larry%20spongebob&client=safari&ved=2ahUKEwi_8Kaa48P7AhU1s2oFHWRrCPMQMygAegUIARDfAQ",
		Banner:      "https://www.google.com/imgres?imgurl=https%3A%2F%2Fstatic.wikia.nocookie.net%2Fspongebob%2Fimages%2F0%2F05%2FLarry_the_Lobster_stock_image_standing.png%2Frevision%2Flatest%3Fcb%3D20220807062551&imgrefurl=https%3A%2F%2Fspongebob.fandom.com%2Fwiki%2FLarry_the_Lobster&tbnid=pbx86w_qC1rc5M&vet=12ahUKEwi_8Kaa48P7AhU1s2oFHWRrCPMQMygAegUIARDfAQ..i&docid=vqku3XZJ7pEJrM&w=1500&h=1500&q=larry%20spongebob&client=safari&ved=2ahUKEwi_8Kaa48P7AhU1s2oFHWRrCPMQMygAegUIARDfAQ",
		Banned:      false,
		Website:     "LarryTheLobstersGym.com",
		Deleted:     false,
	}

	u.Add(is, TableName)
	// this should trigger an error because the user exists already
}

//
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

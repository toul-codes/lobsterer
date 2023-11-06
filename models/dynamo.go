package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
)

const (
	TableName = "Lobsterer-Test"
)

type ItemService struct {
	ItemTable *dynamodb.Client
}

type DynamoConfig struct {
	Region string
	Url    string
	AKID   string
	SAC    string
	ST     string
	Source string
}

func NewItemService(d *DynamoConfig) ItemService {
	dt := CreateLocalClient(d)
	return ItemService{
		ItemTable: dt,
	}
}

func CreateLocalClient(d *DynamoConfig) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(d.Region),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: d.Url}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: d.AKID, SecretAccessKey: d.SAC, SessionToken: d.ST,
				Source: d.Source,
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg)
}

func CreateTableIfNotExists(d *dynamodb.Client, tableName string) {
	if tableExists(d, tableName) {
		log.Printf("table=%v already exists\n", tableName)
		return
	}
	_, err := d.CreateTable(context.TODO(), buildCreateTableInput(tableName))
	if err != nil {
		log.Fatal("CreateTable failed", err)
	}
	log.Printf("created table=%v\n", tableName)
}

func Print(d *dynamodb.Client, tableName string) {
	out, err := d.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		fmt.Printf("Err: %s", err)
	}
	for _, item := range out.Items {
		for k, v := range item {
			fmt.Printf("\n k:%s, v:%s", k, v)
		}
	}
}

func tableExists(d *dynamodb.Client, name string) bool {
	tables, err := d.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatal("ListTables failed", err)
	}
	for _, n := range tables.TableNames {
		if n == name {
			return true
		}
	}
	return false
}

func buildCreateTableInput(tableName string) *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI1PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI1SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI2PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI2SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI3PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI3SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI4PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI4SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI5PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI5SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI1PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI1SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: "ALL",
				},
			},
			{
				IndexName: aws.String("GSI2"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI2PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI2SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: "ALL",
				},
			}, {
				IndexName: aws.String("GSI3"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI3PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI3SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: "ALL",
				},
			}, {
				IndexName: aws.String("GSI4"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI4PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI4SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: "ALL",
				},
			},
			{
				IndexName: aws.String("GSI5"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI5PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI5SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: "ALL",
				},
			},
		},

		TableName:   aws.String(tableName),
		BillingMode: types.BillingModePayPerRequest,
	}
}

func DeleteAllItems(d *dynamodb.Client, tableName string) error {
	log.Printf("deleting all items from table=%v\n", tableName)
	var offset map[string]types.AttributeValue
	for {
		scanInput := &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		}
		if offset != nil {
			scanInput.ExclusiveStartKey = offset
		}
		result, err := d.Scan(context.TODO(), scanInput)
		if err != nil {
			return err
		}

		for _, item := range result.Items {
			_, err := d.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key:       map[string]types.AttributeValue{"PK": item["PK"], "SK": item["SK"]},
			},
			)
			if err != nil {
				return err
			}
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		offset = result.LastEvaluatedKey
	}
	return nil

}

func Delete(d *dynamodb.Client, tableName string) error {
	_, err := d.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		panic(err)
	}
	return nil
}

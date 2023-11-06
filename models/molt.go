package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"math/rand"
	"time"
)

const (
	shardSize = 5
)

type Molt struct {
	ID           string `dynamodbav:"id"`
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	GSI3PK       string `dynamodbav:"GSI3PK"`
	GSI3SK       string `dynamodbav:"GSI3SK"`
	Author       string `dynamodbav:"author"`
	Content      string `dynamodbav:"content"`
	Url          string `dynamodbav:"url"`
	Deleted      bool   `dynamodbav:"deleted"`
	LikeCount    int    `dynamodbav:"like_count"`
	RemoltCount  int    `dynamodbav:"remolt_count"`
	CommentCount int    `dynamodbav:"comment_count"`
}

type Cache struct {
	PK    string `dynamodbav:"PK"`
	SK    string `dynamodbav:"SK"`
	Molts []Molt `dynamodbav:"molts"`
}

type Like struct {
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	GSI4PK string `dynamodbav:"GSI4PK"`
	GSI4SK string `dynamodbav:"GSI4SK"`
}

type Remolt struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	GSI5PK       string `dynamodbav:"GSI5PK"`
	GSI5SK       string `dynamodbav:"GSI5SK"`
	Author       string `dynamodbav:"author"`
	Content      string `dynamodbav:"content"`
	Url          string `dynamodbav:"url"`
	Deleted      bool   `dynamodbav:"deleted"`
	LikeCount    int    `dynamodbav:"like_count"`
	RemoltCount  int    `dynamodbav:"remolt_count"`
	CommentCount int    `dynamodbav:"comment_count"`
}

// FillOcean - is a lambda function that runs every X hour to build the cache
// then that same cache is what each user reads from
// on the latest ocean molts page
// currently builds 5 shards (5 copies of the 25 max latest molts)
func FillOcean(svc ItemService, tablename string) error {
	// retrieve all deals from past day
	l := Latest(svc, tablename)
	fmt.Printf("Length of latest is: %d", len(l))
	for i := 0; i < shardSize; i++ {
		c := &Cache{
			PK:    fmt.Sprintf("MC#%d", i),
			SK:    fmt.Sprintf("MC#%d", i),
			Molts: l,
		}
		item, err := attributevalue.MarshalMap(c)
		if err != nil {
			fmt.Println("ERR: ", err)
			return err
		}
		_, err = svc.ItemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String(tablename),
			Item:      item,
		})
		if err != nil {
			return err
		}

	}
	return nil
}

// Ocean - returns the collection  of the latest molts from a random N shard
func Ocean(svc ItemService, tablename string) []Molt {
	rand.Seed(time.Now().UnixNano())
	cache := rand.Intn(shardSize)
	out, err := svc.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tablename),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("MC#%d", cache)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("MC#%d", cache)},
		},
	})
	log.Printf("\nReading from Cache #: %d", cache)
	if err != nil {
		fmt.Errorf("ERR: %s", err)
	}
	m := make([]Molt, 0)
	molts := out.Item["molts"]
	err = attributevalue.Unmarshal(molts, &m)
	if err != nil {
		fmt.Printf("ERR %s", err)
	}
	return m
}

// Latest - returns all the molts from past day
func Latest(svc ItemService, tablename string) []Molt {
	var limit int32 = 5
	now := time.Now()
	y, mnth, d := now.Date()
	p := dynamodb.NewQueryPaginator(svc.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(tablename),
		Limit:                  aws.Int32(limit), // per pg so get 5 x number requests = amount of molts from today
		IndexName:              aws.String("GSI3"),
		KeyConditionExpression: aws.String("GSI3PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "M#" + fmt.Sprintf("%d-%d-%d", y, int(mnth), d)},
		},
		ScanIndexForward: aws.Bool(false),
	})
	var items []Molt
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var pItems []Molt
		err = attributevalue.UnmarshalListOfMaps(out.Items, &pItems)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		items = append(items, pItems...)

	}
	return items
}

func (u *User) ReMolt(svc ItemService, tablename string, other Molt) error {
	// creates a new molt
	// by passing in the relevent content from other molt
	// increments remolt count on molt
	// increments user's moltCount
	// use the iso 8601 format so that it is easier to query createdAtTime

	// GSI5PK:REPO#<OriginalOwner>#<RepoName>
	// GSI5SK:FORK#<Owner>
	// use the iso 8601 format so that it is easier to query createdAtTime
	m := &Molt{}
	KUID := GenerateKSUID()                   // share one KUID key for time sorting
	m.PK = fmt.Sprintf("M#%s", u.ID)          // M#<UserName>#
	m.SK = fmt.Sprintf("M#%s#%s", u.ID, KUID) // M#<UserName>#<KUID> so molts are users most recent first
	m.Author = u.Display
	m.Content = other.Content

	now := time.Now()
	y, mnth, d := now.Date()

	m.GSI3PK = fmt.Sprintf("M#%s", fmt.Sprintf("%d-%d-%d", y, int(mnth), d))
	m.GSI3SK = fmt.Sprintf("RM#%s", KUID)

	molt, err := attributevalue.MarshalMap(m)

	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}

	tItems := make([]types.TransactWriteItem, 0)
	// delete it from the main table
	tw1 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                molt,
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
			UpdateExpression:    aws.String("set #molt_count = #molt_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#molt_count": "molt_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)

	_, err = svc.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
	}
	return err

}

// CreateMolt - adds molt to db and increments user's MoltCount
func (u *User) CreateMolt(svc ItemService, tablename, content string) error {
	// use the iso 8601 format so that it is easier to query createdAtTime
	m := &Molt{}
	KUID := GenerateKSUID()                   // share one KUID key for time sorting
	m.PK = fmt.Sprintf("M#%s", u.ID)          // M#<UserName>#
	m.SK = fmt.Sprintf("M#%s#%s", u.ID, KUID) // M#<UserName>#<KUID> so molts are users most recent first
	m.Author = u.Display
	m.Content = content

	now := time.Now()
	y, mnth, d := now.Date()

	m.GSI3PK = fmt.Sprintf("M#%s", fmt.Sprintf("%d-%d-%d", y, int(mnth), d))
	m.GSI3SK = fmt.Sprintf("M#%s", KUID)

	molt, err := attributevalue.MarshalMap(m)

	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}

	tItems := make([]types.TransactWriteItem, 0)
	// delete it from the main table
	tw1 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                molt,
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
			UpdateExpression:    aws.String("set #molt_count = #molt_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#molt_count": "molt_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)

	_, err = svc.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
	}
	return err
}

// Trench - create user's trench
// GET /user/:user/trench
func (u *User) Trench(svc ItemService, tablename string) []Molt {
	// have Ocean that is built of sharded data by day
	l := Ocean(svc, tablename)
	following := u.Following(svc, tablename)
	trench := make([]Molt, 0)
	for _, f := range following {
		// use f.SK[2:] b/c only need the ID value not the 'F#' to find user display
		d, err := ByID(f.SK[2:], svc, tablename)
		if err != nil {
			fmt.Errorf("ERR ID LOOKUP: %s", err)
		}
		for _, molt := range l {
			if d.Display == molt.Author && d.Display != u.Display {
				trench = append(trench, molt)
			}
		}
	}
	return trench
}

// Molts - returns the user's latest molts
// / GET /user/:user/molts
func (u *User) Molts(svc ItemService, tablename string) []Molt {
	m := make([]Molt, 0)
	p := dynamodb.NewQueryPaginator(svc.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(tablename),
		Limit:                  aws.Int32(5),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "M#" + u.ID},
		},
		ScanIndexForward: aws.Bool(false), // retrieve users latest molts
	})
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		err = attributevalue.UnmarshalListOfMaps(out.Items, &m)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}

	}
	return m
}

// Like - increments the passed in molt_id's like_count
// POST /molts/like
func (u *User) Like(svc ItemService, tablename string, m Molt) error {
	// this is only for developing mode to work with GIN will need a session
	// which isn't great for running a local test programmatically
	fmt.Printf("\nMID: %s", m.PK[2:])
	fmt.Printf("\nMID: %s", m.SK[2:])
	l := &Like{
		PK:     "ML#" + u.ID,
		SK:     "ML#" + m.PK[2:], // because from front end it will be the full M#N
		GSI4PK: "ML#" + m.PK[2:],
		GSI4SK: "ML#" + u.ID,
	}

	item, err := attributevalue.MarshalMap(l)
	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	tItems := make([]types.TransactWriteItem, 0)
	tw1 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                item,
			ConditionExpression: aws.String("attribute_not_exists(PK)"),
			TableName:           aws.String(tablename),
		},
	}
	//update likes for molt
	tw2 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: m.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: m.SK,
				},
			},
			TableName:           aws.String(tablename),
			ConditionExpression: aws.String("attribute_exists(PK)"),
			UpdateExpression:    aws.String("set #like_count = #like_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#like_count": "like_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)

	_, err = svc.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
	}
	return err
}

func (m *Molt) Delete(svc ItemService, tablename, text string) {
	// sets provided moltu id to deleted=true

}

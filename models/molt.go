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

// FillOcean - is a lambda function that runs every X hour to build the cache
// then that same cache is what each user reads from
// on the latest ocean molts page
// currently builds 5 shards (5 copies of the 25 max latest molts)
func FillOcean(svc ItemService, tablename string) {
	// retrieve all deals from past day
	l := Latest(svc, tablename)
	fmt.Printf("Length of latest is: %d", len(l))
	for i := 0; i < 5; i++ {
		c := &Cache{
			PK:    fmt.Sprintf("MC#%d", i),
			SK:    fmt.Sprintf("MC#%d", i),
			Molts: l,
		}
		item, err := attributevalue.MarshalMap(c)
		if err != nil {
			fmt.Println("ERR: ", err)
			panic(err)
		}
		_, err = svc.ItemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String(tablename),
			Item:      item,
		})

	}
}

// Ocean - returns the collection  of the latest molts from a random N shard
func Ocean(svc ItemService, tablename string) []Molt {
	rand.Seed(time.Now().UnixNano())
	cache := rand.Intn(5)
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

// CreateMolt - adds molt to db and increments user's MoltCount
func (u *User) CreateMolt(svc ItemService, tablename string, content string) error {
	// use the iso 8601 format so that it is easier to query createdAtTime
	m := &Molt{}
	// can use the molts created field to sort globally...?
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
		d, _ := ByID(f.SK[2:], svc, tablename)
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

func (m *Molt) ById(svc ItemService, tablename string, text string) {

}

func (m *Molt) ByAuthor(svc ItemService, tablename string, text string) {

}

func (m *Molt) ByTime(svc ItemService, tablename string, text string) {

}

func (m *Molt) Re(svc ItemService, tablename string, mid string) {
	// (a) Remolt a molt from the ocean view
	// (b) Remalt a molt from the trench view
	// both scenarios require increments the 'original' molts Remolt Count
	// so need to be able to access it from the CachedLatest
	// can do query on GSI3PK (today) where author == cachedAuthor
}

func (m *Molt) Delete(svc ItemService, tablename string, text string) {

}

func (m *Molt) Edit(svc ItemService, tablename string, text string) {

}

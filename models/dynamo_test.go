package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func TestPrint(t *testing.T) {
	is := LocalService()
	Print(is.ItemTable, TableName)
}

func TestDeleteAllItems(t *testing.T) {
	is := LocalService()
	SetUp()
	err := DeleteAllItems(is.ItemTable, TableName)
	if err != nil {
		log.Fatal("failed to delete all items", err)
	}
	scan, err := is.ItemTable.Scan(context.TODO(), &dynamodb.ScanInput{TableName: aws.String(TableName)})
	if err != nil {
		log.Fatal("scan failed", err)
	}
	log.Printf("expected scan to have zero items; it had len=%d\n", len(scan.Items))
	CleanUp()
}

func TestNewItemService(t *testing.T) {
	is := LocalService()
	if is.ItemTable == nil {
		fmt.Printf("error:%v ", is)
	}
}

func TestByID(t *testing.T) {
	is := LocalService()
	SetUp()
	want := strconv.Itoa(rand.Intn(10))
	got, _ := ByID(want, is, TableName)
	if got.ID != want {
		fmt.Errorf("got: %s, want: %s", got.ID, want)
	}
	CleanUp()
}

func TestByName(t *testing.T) {
	is := LocalService()
	SetUp()
	n := strconv.Itoa(rand.Intn(10))
	uByID, _ := ByID(n, is, TableName)
	uByName, _ := ByName(uByID.Display, is, TableName)
	if uByID.ID != uByName.ID {
		fmt.Errorf("byID: %+v, byName: %+v", uByID, uByName)
	}
	if uByID.Display != uByName.Display {
		fmt.Errorf("byID: %+v, byName: %+v", uByID, uByName)
	}
	fmt.Printf("byID: %+v, byName: %+v", uByID, uByName)
	CleanUp()
}

func TestUser_Add(t *testing.T) {
	// adds users
	SetUp()
	CleanUp()
}

func TestUser_Following(t *testing.T) {
	is := LocalService()
	SetUp()
	// get first user
	f, _ := ByID("1", is, TableName)
	// follow second user
	f.Follow(is, TableName, "2")
	l := len(f.Following(is, TableName))
	want := 1
	if l != 1 {
		fmt.Errorf("got: %d, want: %d", l, want)
	}
	CleanUp()
}

func TestUser_Follow(t *testing.T) {
	is := LocalService()
	SetUp()
	f, _ := ByID("1", is, TableName)
	want := 1
	f.Follow(is, TableName, "2")
	s, _ := ByID("2", is, TableName)
	if s.FollowerCount != want {
		fmt.Errorf("got: %d, want: %d", s.FollowerCount, want)
	}
	if f.FollowingCount != 1 {
		fmt.Errorf("got: %d, want: %d", s.FollowerCount, want)
	}
	CleanUp()
}

func TestUser_Unfollow(t *testing.T) {
	is := LocalService()
	SetUp()
	f, _ := ByID("1", is, TableName)
	// follow second user
	f.Follow(is, TableName, "2")
	// change your mind about that
	f.Unfollow(is, TableName, "2")
	s, _ := ByID("2", is, TableName)
	want := 0
	if f.FollowingCount != want {
		fmt.Errorf("got: %d, want: %d", f.FollowingCount, want)
	}
	if s.FollowerCount != want {
		fmt.Errorf("got: %d, want: %d", s.FollowerCount, want)
	}
	CleanUp()
}

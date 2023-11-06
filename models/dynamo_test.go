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
	"time"
)

func TestPrint(t *testing.T) {
	is := LocalService()
	Print(is.ItemTable, TableName)
}

func TestCleanUp(t *testing.T) {
	CleanUp()
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
		fmt.Errorf("\nbyID: %+v, \nbyName: %+v", uByID, uByName)
	}
	if uByID.Display != uByName.Display {
		fmt.Errorf("\nbyID: %+v, \nbyName: %+v", uByID, uByName)
	}
	fmt.Printf("\nbyID: %+v, \nbyName: %+v", uByID, uByName)
	CleanUp()
}

func TestGenerateKSUID(t *testing.T) {
	a := GenerateKSUID()
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(5) // n will be between 0 and 10
	fmt.Printf("Sleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	b := GenerateKSUID()
	if a == b {
		fmt.Errorf("Err: %s, %s", a, b)
	}
	fmt.Printf("\na: %s \nb: %s", a, b)
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
	// follow third user
	f.Follow(is, TableName, "3")
	l := len(f.Following(is, TableName))
	want := 1
	if l != 1 {
		fmt.Errorf("got: %d, want: %d", l, want)
	}

	fmt.Printf("\nf: %v", f)

	fmt.Printf("\nUser is following: %+v", f.Following(is, TableName))
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

func TestUser_Like(t *testing.T) {
	is := LocalService()
	SetUp()
	f1, _ := ByID("1", is, TableName)
	f2, _ := ByID("2", is, TableName)
	f1.CreateMolt(is, TableName, "user 1 is superior because 1 come before 2")
	f2.CreateMolt(is, TableName, "user 1 is inferior because 1 is less than 2")

	l := Latest(is, TableName)
	for _, molt := range l {
		fmt.Printf("\nBefore likes%+v", molt)
	}
	// user one likes the first molt
	f1.Like(is, TableName, l[0])
	// retrieve latest molts again
	l = Latest(is, TableName)
	// print them
	for _, molt := range l {
		fmt.Printf("\nAfter likes %+v", molt)
	}
	CleanUp()
}

func TestUser_CreateMolt(t *testing.T) {
	is := LocalService()
	SetUp()
	f, _ := ByID("1", is, TableName)

	f.CreateMolt(is, TableName, "Hello, ocean trench.")
	want := 1
	if f.MoltCount != want {
		fmt.Errorf("got: %d, want: %d", f.MoltCount, want)
	}
	CleanUp()
}

func TestUser_Molts(t *testing.T) {
	is := LocalService()
	SetUp()
	f, _ := ByID("1", is, TableName)

	f.CreateMolt(is, TableName, "Hello, ocean trench.")
	r := f.Molts(is, TableName)
	if len(r) != 1 {
		fmt.Printf("got: %d, Want: %d", len(r), 1)
	}
	fmt.Printf("\n\n%+v", r)
	CleanUp()
}

func TestUser_ReMolt(t *testing.T) {
	is := LocalService()
	SetUp()
	f, _ := ByID("1", is, TableName)
	f.CreateMolt(is, TableName, "I'm the first crab...")
	f2, _ := ByID("2", is, TableName)
	l := Latest(is, TableName)
	f2.ReMolt(is, TableName, l[0])
	l = Latest(is, TableName)
	fmt.Printf("\nLatest: %v", l)
	m := f2.Molts(is, TableName)
	fmt.Printf("\n user 2 molts: %v", m)
	CleanUp()
}

func TestUser_Trench(t *testing.T) {
	is := LocalService()
	SetUp()
	f1, _ := ByID("1", is, TableName)
	f2, _ := ByID("2", is, TableName)
	f3, _ := ByID("3", is, TableName)
	f4, _ := ByID("4", is, TableName)
	f5, _ := ByID("5", is, TableName)
	//
	rand.Seed(time.Now().UnixNano())
	n := 3 // n will be between 0 and 10
	f1.CreateMolt(is, TableName, "User 1 has molted")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "User 2 has molted")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "user 3 has molted")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "user 4 has molted")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "user 5 has molted")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "user 1 likes pizza")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "user 2 just did a #2")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "user 3 doesn't like hot pockets")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "user 4 likes dogs")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "user 5 likes ants")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "user 1 said 'damn that abbot'")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "use 2 saw a duck")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "user 3 saw a rabbit")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "user 4 doesn't like the word f-o-u-r")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "user 5 is a biden bro")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "user 1 didn't do their taxes ")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "user 2 molts too much")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "user 3 doesn't molt enough")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "user 4 likes squares tho")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "user 5 doesn't like to drink tea")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "user 1 is thinking GO is a pain for this")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "user 2 eats their boogers")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "user 3 likes harry potter")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "user 4 likes reddit")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "user 5 ate a person alive")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "1")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "user 2 in unhinged")

	FillOcean(is, TableName) // caching
	// follow two users
	f1.Follow(is, TableName, "3")
	f1.Follow(is, TableName, "5")
	// now, reading the trench should show the ones user 1 is following
	res := f1.Trench(is, TableName)
	for _, m := range res {
		fmt.Printf("\n%+v", m)
	}
	CleanUp()
}

func TestFillOcean(t *testing.T) {
	is := LocalService()
	SetUp()
	f1, _ := ByID("1", is, TableName)
	f2, _ := ByID("2", is, TableName)
	f3, _ := ByID("3", is, TableName)
	f4, _ := ByID("4", is, TableName)
	f5, _ := ByID("5", is, TableName)
	//
	rand.Seed(time.Now().UnixNano())
	n := 3
	f1.CreateMolt(is, TableName, "1vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "2vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "3vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "4vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "5vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "6vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "7vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "8vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "9vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "10vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "11vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "12h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "13h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "14h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "15h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "16h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "17h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "18h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "19h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "20")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "1h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "21")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "22")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "23")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "24")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "1")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "25")
	err := FillOcean(is, TableName) // caching
	if err != nil {
		t.Errorf("%s", err)
	}
	CleanUp()
}

func TestOcean(t *testing.T) {
	is := LocalService()
	SetUp()
	f1, _ := ByID("1", is, TableName)
	f2, _ := ByID("2", is, TableName)
	f3, _ := ByID("3", is, TableName)
	f4, _ := ByID("4", is, TableName)
	f5, _ := ByID("5", is, TableName)
	//
	rand.Seed(time.Now().UnixNano())
	n := 3
	f1.CreateMolt(is, TableName, "1vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "2vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "3vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "4vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "5vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "6vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "7vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "8vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "9vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "10vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "11vh")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "12h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "13h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "14h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "15h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "16h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "17h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "18h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "19h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "20")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "1h")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "21")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "22")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "23")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "24")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "1")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "25")
	err := FillOcean(is, TableName) // caching
	if err != nil {
		t.Errorf("%s", err)
	}
	latest := Ocean(is, TableName)
	for _, molt := range latest {
		fmt.Printf("\nm: %+v", molt)
	}
	CleanUp()
}

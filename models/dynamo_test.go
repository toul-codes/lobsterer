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
	fmt.Printf("User is following: %+v", f.Following(is, TableName))
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

func TestCreateMolt(t *testing.T) {
	is := LocalService()
	SetUp()
	f, _ := ByID("1", is, TableName)
	// follow second user
	f.CreateMolt(is, TableName, "Hello, ocean trench.")
	want := 1
	if f.MoltCount != want {
		fmt.Errorf("got: %d, want: %d", f.MoltCount, want)
	}
	CleanUp()
}

func TestUser_CreateMolt(t *testing.T) {
	is := LocalService()
	SetUp()
	f, _ := ByID("1", is, TableName)
	// follow second user
	f.CreateMolt(is, TableName, "Hello, ocean trench.")
	want := 1
	if f.MoltCount != want {
		fmt.Errorf("got: %d, want: %d", f.MoltCount, want)
	}
	r := f.Molts(is, TableName)
	fmt.Printf("\n\n%+v", r)
	CleanUp()
}

func TestLatest(t *testing.T) {
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
	f1.CreateMolt(is, TableName, "1vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "2vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "3vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "4vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "5vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "6vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "7vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "8vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "9vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "10vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "11vhlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "12hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "13hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "14hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "15hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "16hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "17hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "18hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "19hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "20lkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "1hlkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "21lkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f3.CreateMolt(is, TableName, "22lkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f4.CreateMolt(is, TableName, "23lkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f5.CreateMolt(is, TableName, "24lkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f1.CreateMolt(is, TableName, "1lkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	fmt.Printf("\nSleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
	f2.CreateMolt(is, TableName, "25lkbcrqbnaiqbsxfnuhmxszgzzazqlekimwegglwsqcubixlxvexlmulgehefnqhzuzelclmctrjtstzbndnadnztagmxebzbrdgtwmpdmmljajjfsoqbzwcohnwnsiztnxuuaggotqmencxnudsfvtzbfsxwswjbgwjbwqnajucuphtnmqjhwexptdjvwnrifixeivaesycopelxrmempnmfmebnlgdipbiqiscmrychwmfahleysstvvqqekoftiregepoaecdfrvlgykirkcinwwtitsgsmlqyvvcfwtqvtrkyxpjapumojopjotjnpczxbonyemqkdlyrwkgopbmwwnmknzqiynxwvvtoltydorkygparytpnhrbxwktpiklhytwqpbyyvdfkvrwackgewsdfuuxdgldxqemynlvtzldsxsipsmdavmiokamkymogcuyhlqnthamimeusbmuhnjsxuldmikvnwvwsbujllqgfqelzzifhfquqqtbfbj")
	f1.Like(is, TableName, "M#3")
	FillOcean(is, TableName)

	f1.Follow(is, TableName, "3")

	res := f1.Trench(is, TableName)
	for _, m := range res {
		fmt.Printf("\n%+v", m)
	}

	CleanUp()
}

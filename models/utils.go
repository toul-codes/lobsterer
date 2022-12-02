package models

import (
	"fmt"
	"math/rand"
	"strconv"
)

// RandStringRunes - creates random strings
func RandStringRunes(n int, runes []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

// CleanUp - Deletes table so that each test uses a fresh one
func CleanUp() {
	is := LocalService()
	err := Delete(is.ItemTable, TableName)
	if err != nil {
		fmt.Printf("table delete err: %s", err)
	}
}

// TestSetUp - is used for tests and requires Local DynamoDB docker container to be running
func TestSetUp() {
	charRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	is := LocalService()
	// Delete table

	// docker should be running as well as the container
	CreateTableIfNotExists(is.ItemTable, TableName)

	// Create 10 users in db table
	// will keep this small for faster test times
	for i := 0; i < 10; i++ {
		u := User{
			ID:             strconv.Itoa(i),
			Name:           RandStringRunes(10, charRunes),
			Email:          "",
			Display:        "",
			Description:    "",
			Verified:       false,
			Avatar:         "",
			Banner:         "",
			Banned:         false,
			Website:        "",
			Deleted:        false,
			FollowerCount:  0,
			FollowingCount: 0,
		}
		// add them to the table
		u.Add(is, TableName)
	}

}

// LocalService - Used for tdd tests
func LocalService() ItemService {
	return NewItemService(&DynamoConfig{
		Region: "us-west-2",
		Url:    "http://localhost:8000",
		AKID:   "getGudKid",
		SAC:    "eatMorCrabs",
		ST:     "thisissuchasecret",
		Source: "noneofthismattersitsalllocalyfake",
	})
}

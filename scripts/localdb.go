package main

import "github.com/toul-codes/lobsterer/models"

func main() {
	is := models.LocalService()
	models.CreateTableIfNotExists(is.ItemTable, models.TableName)
}

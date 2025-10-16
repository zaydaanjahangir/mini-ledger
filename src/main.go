package main

import (
	"mini-ledger/src/db"
)

func main(){
	db := db.InitDB()
	defer db.Close()
}
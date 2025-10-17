package main

import (
	db "mini-ledger/src/db"
	api "mini-ledger/src/api"
)

func main(){
	db := db.InitDB()
	defer db.Close()
	api.StartServer(db)
}
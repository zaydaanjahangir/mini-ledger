package main

import (
	api "mini-ledger/src/api"
	db "mini-ledger/src/db"
	ledger "mini-ledger/src/ledger"
)

func main(){
	db := db.InitDB()
	defer db.Close()
	go ledger.StartWorker(db, 1000)
	api.StartServer(db)
}

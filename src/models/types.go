package models

type Account struct {
    Name string `json:"name"`
}

type Entry struct {
	Account string  `json:"account"`
	Amount  int64 `json:"amount"`
}

type Transaction struct {
	Description string `json:"description"`
	Entries     []Entry `json:"entries"`
}
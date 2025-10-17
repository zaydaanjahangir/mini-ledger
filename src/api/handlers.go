package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"log"
)

type Account struct {
    Name string `json:"name"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func StartServer(db *sql.DB) {
	mux := http.NewServeMux()

	mux.HandleFunc("/accounts", accountHandler(db))

	http.ListenAndServe(":8080", mux)
}

func accountHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request){
		if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
		}

		var account Account
		if err := json.NewDecoder(req.Body).Decode(&account); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO accounts (name) VALUES (?)", account.Name)
		if err != nil{
			http.Error(w, "Error creating account", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Account created successfully",
		})
	}

}


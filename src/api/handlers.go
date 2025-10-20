package handlers

import (
	"database/sql"
	"encoding/json"
	"mini-ledger/src/ledger"
	"mini-ledger/src/models"
	"net/http"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func StartServer(db *sql.DB) {
	mux := http.NewServeMux()

	mux.HandleFunc("/accounts", accountHandler(db))
	mux.HandleFunc("/transactions", transactionHandler(db))

	http.ListenAndServe(":8080", mux)
}

func accountHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request){
		if req.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var account models.Account
		if err := json.NewDecoder(req.Body).Decode(&account); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		err := ledger.PostAccount(db, account)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Status:  "success",
			Message: "Account created successfully",
		})
	}

}

func transactionHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var transaction models.Transaction
		if err := json.NewDecoder(req.Body).Decode(&transaction); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if len(transaction.Entries) < 2 {
			http.Error(w, "Transaction must have at least two entries", http.StatusBadRequest)
			return
		}

		err := ledger.PostTransaction(db, transaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Transaction validated successfully",
		})

	}
}



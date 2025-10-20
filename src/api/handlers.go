package handlers

import (
	"database/sql"
	"encoding/json"
	"encoding/hex"
	"net/http"
	// "log"
	"fmt"
	"crypto/sha256"
    "time"
)

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

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func hashTransaction(transaction Transaction) string{
	data, _ := json.Marshal(transaction)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
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

		var transaction Transaction
		if err := json.NewDecoder(req.Body).Decode(&transaction); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if len(transaction.Entries) < 2 {
			http.Error(w, "Transaction must have at least two entries", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Error starting transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// TODO: Store each account’s ID in a map so you don’t query twice
		var total int64
		for _, entry := range transaction.Entries {
			var exists bool
			err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE name = ?)", entry.Account).Scan(&exists)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, fmt.Sprintf("Account %s does not exist", entry.Account), http.StatusBadRequest)
				return
			}

			total += entry.Amount
		}

		if total != 0 {
			http.Error(w, "Entries do not balance (sum must equal 0)", http.StatusBadRequest)
			return
		}

		res, err := tx.Exec("INSERT INTO transactions(description) VALUES(?)", transaction.Description)
		if err != nil{
			http.Error(w, "Error fetching transaction id", http.StatusInternalServerError)
			return
		}
		transactionID, _ := res.LastInsertId()
		

		
		stmt, err := tx.Prepare("INSERT INTO entries(transaction_id, account_id, amount) VALUES(?, ? ,?)")
		if err != nil {
			http.Error(w, "Error preparing query", http.StatusInternalServerError)
		}
		defer stmt.Close()
		for _, entry := range transaction.Entries {
			var accountID int
			err = tx.QueryRow("SELECT id FROM accounts WHERE name = ?", entry.Account).Scan(&accountID)
			if err != nil {
				http.Error(w, "Error finding account id", http.StatusInternalServerError)
				return
			}
			_, err = stmt.Exec(transactionID, accountID, entry.Amount)
			if err != nil {
				http.Error(w, "Error writing entries", http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, "Error committing transaction", http.StatusInternalServerError)
			return
		}

		// Log transactions for verifiability
		txHash := hashTransaction(transaction)
		_, err = db.Exec("INSERT INTO transaction_hashes(hash, created_at) VALUES (?, ?)", txHash, time.Now())
		if err != nil {
			http.Error(w, "Error storing tx hash:", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Transaction validated successfully",
		})

	}
}



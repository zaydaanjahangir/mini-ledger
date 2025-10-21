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
	mux.HandleFunc("/accounts/all", getAccountsHandler(db))
	mux.HandleFunc("/transactions/all", getTransactionsHandler(db))
	mux.HandleFunc("/verify", verifyHandler(db))

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

func getAccountsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`
			SELECT 
				a.name AS account,
				COALESCE(SUM(e.amount), 0) AS balance
			FROM accounts a
			LEFT JOIN entries e ON a.id = e.account_id
			GROUP BY a.id
			ORDER BY a.name;
		`)
		if err != nil {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type AccountBalance struct {
			Account string `json:"account"`
			Balance int64  `json:"balance"`
		}

		var results []AccountBalance
		for rows.Next() {
			var ab AccountBalance
			if err := rows.Scan(&ab.Account, &ab.Balance); err != nil {
				http.Error(w, "Error scanning row", http.StatusInternalServerError)
				return
			}
			results = append(results, ab)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func getTransactionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := `
			SELECT 
				t.id,
				t.description,
				t.created_at,
				a.name,
				e.amount
			FROM transactions t
			JOIN entries e ON t.id = e.transaction_id
			JOIN accounts a ON e.account_id = a.id
			ORDER BY t.created_at DESC, t.id;
		`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type EntryView struct {
			Account string `json:"account"`
			Amount  int64  `json:"amount"`
		}
		type TransactionView struct {
			ID          int64        `json:"id"`
			Description string       `json:"description"`
			CreatedAt   string       `json:"created_at"`
			Entries     []EntryView  `json:"entries"`
		}

		txMap := make(map[int64]*TransactionView)
		for rows.Next() {
			var id int64
			var desc, createdAt, account string
			var amount int64
			if err := rows.Scan(&id, &desc, &createdAt, &account, &amount); err != nil {
				http.Error(w, "Error scanning rows", http.StatusInternalServerError)
				return
			}
			entry := EntryView{Account: account, Amount: amount}
			if txMap[id] == nil {
				txMap[id] = &TransactionView{
					ID:          id,
					Description: desc,
					CreatedAt:   createdAt,
					Entries:     []EntryView{},
				}
			}
			txMap[id].Entries = append(txMap[id].Entries, entry)
		}

		var result []TransactionView
		for _, tx := range txMap {
			result = append(result, *tx)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func verifyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ok := ledger.VerifyDigests(db)

		status := "valid"
		if !ok {
			status = "invalid"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": status,
		})
	}
}



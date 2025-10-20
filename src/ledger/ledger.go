package ledger

import (
	"database/sql"
	"mini-ledger/src/models"
	"time"
)

func PostAccount(db *sql.DB, account models.Account) error {
	_, err := db.Exec("INSERT INTO accounts (name) VALUES (?)", account.Name)
		if err != nil{
			return err
		}
	return nil
}

func PostTransaction(db *sql.DB, transaction models.Transaction) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// TODO: Store each account’s ID in a map so you don’t query twice
		var total int64
		for _, entry := range transaction.Entries {
			var exists bool
			err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE name = ?)", entry.Account).Scan(&exists)
			if err != nil {
				return err
			}
			if !exists {
				return err
			}

			total += entry.Amount
		}

		if total != 0 {
			return err
		}

		res, err := tx.Exec("INSERT INTO transactions(description) VALUES(?)", transaction.Description)
		if err != nil{
			return err
		}
		transactionID, _ := res.LastInsertId()
		

		
		stmt, err := tx.Prepare("INSERT INTO entries(transaction_id, account_id, amount) VALUES(?, ? ,?)")
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, entry := range transaction.Entries {
			var accountID int
			err = tx.QueryRow("SELECT id FROM accounts WHERE name = ?", entry.Account).Scan(&accountID)
			if err != nil {
				return err
			}
			_, err = stmt.Exec(transactionID, accountID, entry.Amount)
			if err != nil {
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		// Log transactions for verifiability
		txHash := HashTransaction(transaction)
		_, err = db.Exec("INSERT INTO transaction_hashes(hash, created_at) VALUES (?, ?)", txHash, time.Now())
		if err != nil {
			return err
		}

		MaybeProduceDigest(db, 10)
		return nil
}